# Go 内存管理

## 1. 概述

Go 的内存分配器基于 **TCMalloc（Thread-Caching Malloc）** 设计思想，核心目标是：

- **减少锁竞争**：通过多级缓存（mcache → mcentral → mheap），小对象分配几乎无锁
- **减少内存碎片**：使用 span 机制按固定大小分级管理内存
- **高效利用**：空闲内存复用，避免频繁向操作系统申请/释放

### 内存分配层级

```
  goroutine 申请内存
        │
        ▼
  ┌───────────┐
  │  mcache   │  每个 P 私有，无锁
  │ (per-P)   │
  └─────┬─────┘
        │ 本级缓存不足
        ▼
  ┌───────────┐
  │ mcentral  │  全局共享，有锁（每个 sizeclass 一个）
  │           │
  └─────┬─────┘
        │ 本级缓存不足
        ▼
  ┌───────────┐
  │  mheap    │  全局唯一，管理整个堆内存
  │           │
  └─────┬─────┘
        │ 堆内存不足
        ▼
  ┌───────────┐
  │    OS     │  通过 mmap 等系统调用向操作系统申请
  └───────────┘
```

## 2. 核心概念

### 2.1 内存页（Page）

Go 运行时中的内存页大小为 **8KB**（非操作系统的 4KB 页），是内存管理的最小分配单元。

### 2.2 mspan

`mspan` 是 Go 内存管理的基本单元，是一组**连续的内存页**（Page），被划分为固定大小的若干个 `object`。

```
                         mspan
┌──────────────────────────────────────────┐
│  page  │  page  │  page  │  page  │ ...  │
│        │        │        │        │      │
│ ┌────┐ │ ┌────┐ │ ┌────┐ │ ┌────┐ │      │
│ │obj │ │ │obj │ │ │obj │ │ │obj │ │      │
│ ├────┤ │ ├────┤ │ ├────┤ │ ├────┤ │      │
│ │obj │ │ │obj │ │ │obj │ │ │obj │ │      │
│ ├────┤ │ ├────┤ │ ├────┤ │ ├────┤ │      │
│ │obj │ │ │obj │ │ │obj │ │ │obj │ │      │
│ └────┘ │ └────┘ │ └────┘ │ └────┘ │      │
└──────────────────────────────────────────┘
          npages 个页，每个 object 大小相同
```

```go
// src/runtime/mheap.go

type mspan struct {
    next     *mspan     // 链表中的下一个 span
    prev     *mspan     // 链表中的上一个 span
    startAddr uintptr   // 起始地址
    npages    uintptr   // 页数
    spanclass spanClass // sizeclass 和 noscan 标记
    allocBits  *gcBits  // 分配位图（标记哪些 object 已分配）
    gcmarkBits *gcBits  // GC 标记位图
    // ...
}
```

### 2.3 Size Class（大小分级）

Go 将小对象按大小分为 **67 个 size class**（加上 class 0 用于大对象，共 68 个）。每个 class 对应一个固定的 object 大小。

| class | 字节数 | 每个 span 的 object 数 | 尾部浪费 |
|-------|--------|----------------------|----------|
| 1 | 8 | 1024 | 0 |
| 2 | 16 | 512 | 0 |
| 3 | 24 | 341 | 0 |
| 4 | 32 | 256 | 0 |
| 5 | 48 | 170 | 32 |
| ... | ... | ... | ... |
| 66 | 28672 | 2 | 0 |
| 67 | 32768 | 1 | 0 |

分配策略：为对象找到**最小的能容纳它的 size class**。例如申请 20 字节 → 分配到 class 3（24 字节），内部碎片 4 字节。

### 2.4 spanClass

每个 size class 实际上对应**两个** `spanClass`：

- **scan**（包含指针的对象）：GC 需要扫描其中的指针
- **noscan**（不含指针的对象）：GC 可以跳过扫描

```
spanClass = sizeclass << 1 | noscan
```

这样 67 个 size class 变为 134 个 spanClass，noscan 的 span 在 GC 时可以直接跳过，减少扫描开销。

## 3. 分配器组件

### 3.1 mcache（每个 P 的本地缓存）

每个 P 拥有一个 `mcache`，缓存了各个 spanClass 的 `mspan`。

```go
// src/runtime/mcache.go

type mcache struct {
    alloc [numSpanClasses]*mspan // 每个 spanClass 一个 mspan（共 134 个）
    tiny       uintptr           // 微对象分配器的当前块
    tinyoffset uintptr           // tiny 块内的偏移
    tinyAllocs uintptr           // 微对象分配计数
    // ...
}
```

**优势**：mcache 是 P 私有的，分配时无需加锁。

分配流程：
1. 从 `mcache.alloc[spanClass]` 中找到对应的 mspan
2. 在 mspan 的 `allocBits` 中查找空闲 object
3. 若 mspan 已满，从 mcentral 获取新的 mspan 替换

### 3.2 mcentral（全局中心缓存）

每个 spanClass 对应一个 `mcentral`，维护该 spanClass 的所有 mspan。

```go
// src/runtime/mcentral.go

type mcentral struct {
    spanclass spanClass

    // Go 1.16+ 使用两个 spanSet 替代之前的链表
    partial [2]spanSet // 有空闲 object 的 mspan 集合
    full    [2]spanSet // 无空闲 object 的 mspan 集合
}
```

`partial` 和 `full` 各有 2 个，分别对应 `sweepgen` 的奇偶（GC 的已清扫/未清扫 span）。

当 mcache 向 mcentral 申请 mspan 时：
1. 从 `partial` 中取一个有空闲 object 的 mspan 给 mcache
2. 将 mcache 归还的满 mspan 放入 `full`
3. 若 `partial` 也为空，向 mheap 申请新的内存页创建 mspan

### 3.3 mheap（全局堆）

`mheap` 是全局唯一的，管理整个 Go 堆内存。

```go
// src/runtime/mheap.go

type mheap struct {
    lock      mutex
    pages     pageAlloc  // 页分配器（基数树）
    allspans  []*mspan   // 所有创建过的 mspan
    central   [numSpanClasses]struct {
        mcentral mcentral
    }
    // ...
}
```

职责：
- 管理大块内存（通过 `mmap` 向 OS 申请）
- 为 mcentral 提供 mspan
- 大对象（>32KB）直接在 mheap 上分配

## 4. 三类对象的分配策略

Go 将对象按大小分为三类，采用不同的分配路径：

```
         对象大小
            │
    ┌───────┼──────────┐
    │       │          │
 ≤ 16B   17B-32KB   > 32KB
    │       │          │
  微对象   小对象     大对象
    │       │          │
    ▼       ▼          ▼
  tiny    mcache     mheap
 allocator  │      （直接分配）
    │       │
    ▼       ▼
  mcache  mcache
```

### 4.1 微对象（Tiny Object，≤ 16B 且不含指针）

对于不含指针、大小 ≤ 16 字节的对象（如 `bool`、`int8`、短字符串等），使用 **tiny allocator** 将多个微对象合并到同一个 16 字节的内存块中。

```
  tiny block (16 bytes)
  ┌──┬───┬──┬─────────┐
  │b │i8 │b │  空闲    │
  │1B│1B │1B│  13B    │
  └──┴───┴──┴─────────┘
  ↑            ↑
  tinyAddr     tinyoffset = 3
```

优势：大幅减少小对象的内存分配次数和内存浪费。

### 4.2 小对象（Small Object，17B - 32KB）

1. 计算对象大小对应的 size class
2. 从当前 P 的 `mcache.alloc[spanClass]` 获取 mspan
3. 在 mspan 中通过 `allocBits` 位图查找空闲 object，使用 `ctz`（count trailing zeros）快速定位
4. 若 mspan 满了：mcache → mcentral → mheap → OS 逐级申请

### 4.3 大对象（Large Object，> 32KB）

直接在 `mheap` 上分配，绕过 mcache 和 mcentral。分配对应大小的连续页（以 8KB 对齐）。

## 5. 栈内存管理

Go 的 goroutine 栈和堆内存管理是分开的。

### 5.1 栈增长

- goroutine 初始栈大小 **2KB**
- 每次函数调用时，编译器插入 **栈溢出检查**（`morestack`）
- 栈不够时，分配一个**两倍大小**的新栈，将旧栈内容拷贝过去（**连续栈**，Go 1.4+）
- 旧栈中的指针会被更新为指向新栈的地址

### 5.2 栈收缩

- GC 时检查：如果栈使用量不足当前栈大小的 **1/4**，则栈减半
- 栈的最小值为 2KB，不会再缩小

### 5.3 栈 vs 堆分配

编译器通过**逃逸分析**决定变量分配在栈上还是堆上。详见 [内存逃逸](./escape_to_heap.md)。

- 栈分配：函数返回时自动回收，零开销
- 堆分配：需要 GC 回收，有额外开销

## 6. 内存对齐

Go 编译器会对结构体字段进行**内存对齐**，保证 CPU 访问效率。

### 6.1 对齐规则

| 类型 | 对齐值 |
|------|--------|
| `bool` / `int8` / `uint8` / `byte` | 1 字节 |
| `int16` / `uint16` | 2 字节 |
| `int32` / `uint32` / `float32` | 4 字节 |
| `int64` / `uint64` / `float64` / `pointer` | 8 字节（64 位系统） |

结构体整体大小会向最大字段的对齐值对齐。

### 6.2 字段顺序影响大小

```go
// 占 24 字节（有 padding）
type Bad struct {
    a bool    // 1 字节 + 7 字节 padding
    b int64   // 8 字节
    c bool    // 1 字节 + 7 字节 padding
}

// 占 16 字节（紧凑排列）
type Good struct {
    b int64   // 8 字节
    a bool    // 1 字节
    c bool    // 1 字节 + 6 字节 padding
}
```

使用 `unsafe.Sizeof()` 和 `unsafe.Alignof()` 可查看大小和对齐值。可以使用 `fieldalignment` 工具自动优化：

```bash
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
fieldalignment -fix ./...
```

## 7. 虚拟内存布局

Go 1.11+ 使用**稀疏堆**（Sparse Heap）管理虚拟内存，不再使用连续的线性地址空间。

堆内存由多个 **heapArena** 组成，每个 arena 大小：
- Linux/64 位：**64MB**
- Windows/64 位：**4MB**

```go
// src/runtime/mheap.go

type heapArena struct {
    bitmap     [heapArenaBitmapWords]uintptr // GC 位图（标记指针位置）
    spans      [pagesPerArena]*mspan         // 页到 mspan 的映射
    pageInUse  [pagesPerArena / 8]uint8      // 页使用状态
    pageMarks  [pagesPerArena / 8]uint8      // GC 标记状态
    // ...
}
```

`mheap` 通过 `arenas` 数组（二级索引）管理所有 `heapArena`。

## 8. 内存归还操作系统

Go 运行时并不会立即将释放的内存归还 OS，而是标记为可回收：

| 操作 | 说明 |
|------|------|
| `madvise(MADV_FREE)` | 告诉 OS 这些页可以回收（Linux 4.5+），但 OS 在内存紧张时才真正回收 |
| `madvise(MADV_DONTNEED)` | 立即释放物理内存（较老的 Linux 内核） |
| `runtime/debug.FreeOSMemory()` | 手动强制归还内存给 OS |

`scavenger`（清道夫）goroutine 在后台持续运行，逐步将长期不用的内存归还给 OS。

可通过环境变量 `GODEBUG=madvdontneed=1` 强制使用 `MADV_DONTNEED` 策略。

## 9. 完整分配流程

```
  new(T) / make(T) / 字面量
           │
           ▼
    ┌──────────────┐
    │  逃逸分析     │
    │ 编译期决定    │
    └──┬───────┬───┘
       │       │
    不逃逸    逃逸
       │       │
       ▼       ▼
     栈分配   堆分配
              │
       ┌──────┼──────────┐
       │      │          │
    ≤ 16B  17B-32KB   > 32KB
       │      │          │
       ▼      ▼          ▼
     tiny   mcache     mheap
   allocator  │      直接分配
       │      ▼
       ▼   allocBits
    合并到  定位空闲 obj
   16B 块     │
       │   span 满？
       │   ├── 否 → 返回 obj
       │   └── 是 ↓
       │      mcentral
       │   partial 取 span
       │      │
       │   partial 空？
       │   ├── 否 → 返回 span
       │   └── 是 ↓
       │      mheap
       │   分配连续页
       │      │
       │   pages 不足？
       │   ├── 否 → 创建 span
       │   └── 是 ↓
       │      OS (mmap)
       │   申请新的 arena
       │
       ▼
    返回内存地址
```

## 10. 常用诊断工具

| 工具/方法 | 用途 |
|-----------|------|
| `runtime.MemStats` | 获取详细的内存统计信息 |
| `runtime.ReadMemStats(&ms)` | 读取当前内存状态 |
| `pprof` (`heap` profile) | 查看堆内存分配热点 |
| `go build -gcflags='-m'` | 查看逃逸分析结果 |
| `GODEBUG=gctrace=1` | 打印 GC 日志 |
| `go tool trace` | 可视化运行时 trace |

```go
var ms runtime.MemStats
runtime.ReadMemStats(&ms)
fmt.Printf("HeapAlloc = %d MB\n", ms.HeapAlloc/1024/1024)   // 堆上已分配的内存
fmt.Printf("HeapSys   = %d MB\n", ms.HeapSys/1024/1024)     // 从 OS 获取的堆内存
fmt.Printf("HeapIdle  = %d MB\n", ms.HeapIdle/1024/1024)    // 空闲的堆内存
fmt.Printf("NumGC     = %d\n", ms.NumGC)                    // GC 次数
```

## 参考资料

- [Go 内存分配器源码](https://github.com/golang/go/blob/master/src/runtime/malloc.go)
- [TCMalloc: Thread-Caching Malloc](https://google.github.io/tcmalloc/design.html)
- [A visual guide to Go Memory Allocator from scratch](https://medium.com/@ankur_anand/a-visual-guide-to-golang-memory-allocator-from-ground-up-e132f94d96b4)
- [Go Memory Management and Allocation](https://medium.com/a-journey-with-go/go-memory-management-and-allocation-a7396d430f44)
