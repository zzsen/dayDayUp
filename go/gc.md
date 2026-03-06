# Go GC（垃圾回收）

## 1. 概述

Go 采用 **并发、三色标记、清扫** 的垃圾回收算法，从 Go 1.5 开始逐步演进为低延迟的并发 GC。

| 版本 | 里程碑 |
|------|--------|
| Go 1.1 | STW 标记清扫（全程暂停） |
| Go 1.3 | 标记阶段并发化，清扫阶段并发化 |
| Go 1.5 | 三色标记 + 写屏障，大幅降低 STW 时间 |
| Go 1.8 | 混合写屏障（Hybrid Write Barrier），STW 降至亚毫秒级 |
| Go 1.12 | 改进清扫器性能 |
| Go 1.19 | 引入 `runtime/debug.SetMemoryLimit`（Soft Memory Limit） |

当前 Go 的 GC 目标：**低延迟**（STW 时间通常 < 1ms），而非最大吞吐量。

### STW（Stop The World）

STW 指 GC 在某些阶段**暂停所有用户 goroutine**，整个程序"停下来"，只有 GC 在工作。Go 早期（1.1）的 GC 全程 STW，暂停可达数百毫秒。当前版本仅在两个极短阶段需要 STW，中间的并发标记和并发清扫阶段用户代码正常运行，不暂停。降低 STW 时间是 Go GC 历代优化的核心目标。

| STW 阶段 | 做什么 | 典型耗时 |
|----------|--------|----------|
| Mark Setup（标记准备） | 开启写屏障、将所有栈标记为"需要扫描" | 10-30μs |
| Mark Termination（标记终止） | 确认所有标记工作完成、关闭写屏障 | 10-30μs |

## 2. 三色标记法

三色标记将堆上的对象分为三种颜色：

| 颜色 | 含义 |
|------|------|
| **白色** | 未被访问的对象（GC 结束后，白色对象被视为垃圾回收） |
| **灰色** | 已被访问，但其引用的对象尚未全部扫描 |
| **黑色** | 已被访问，且其引用的对象全部已扫描 |

> **扫描（scan）** 指遍历一个对象内部的所有指针字段，找出它引用了哪些其他对象。例如一个 `Order` 结构体包含 `*User` 和 `[]*Item` 两个指针字段，扫描 `Order` 就是检查这些指针指向了哪些对象，并将它们标记为灰色。扫描完成后，`Order` 从灰色变为黑色。

### 2.1 标记流程

```
初始状态：所有对象为白色

步骤 1：将 GC Root（栈、全局变量等）直接引用的对象标记为灰色

步骤 2：从灰色集合中取出一个对象
       → 将它标记为黑色
       → 将它引用的所有白色对象标记为灰色

步骤 3：重复步骤 2，直到灰色集合为空

步骤 4：回收所有白色对象
```

图示：

```
  GC Root
    │
    ▼
  ┌───┐     ┌───┐     ┌───┐
  │ A │────►│ B │────►│ C │       初始：全白
  └───┘     └───┘     └───┘
                       ┌───┐
                       │ D │       D 不可达
                       └───┘

  第 1 步: A → 灰色
  第 2 步: A → 黑色, B → 灰色
  第 3 步: B → 黑色, C → 灰色
  第 4 步: C → 黑色
  结束: D 仍为白色 → 回收 D
```

### 2.2 并发标记的问题

三色标记在 STW 环境下是安全的。但 Go 的 GC 是**并发**的——标记阶段 mutator（用户代码）仍在运行。这可能导致**漏标**：

**漏标条件**（同时满足时发生）：
1. 黑色对象引用了白色对象（黑色对象新增了对白色对象的引用）
2. 所有灰色对象到该白色对象的引用路径被断开

漏标示例：

```
初始状态：A（黑色，已扫描完），B（灰色），C（白色）
  A          B ────→ C
             灰色引用着白色 C，C 本应在 B 被扫描时变灰

此时用户代码并发执行了两步操作：
  ① A.ref = C     （黑色 A 新增了对白色 C 的引用 → 满足条件 1）
  ② B.ref = nil   （灰色 B 断开了对 C 的引用   → 满足条件 2）

结果：
  A（黑色）──→ C（白色）     B（灰色）──→ nil
  A 已经是黑色，不会再被扫描，GC 不会发现 A→C 这条引用
  B 扫描时发现没有引用任何对象
  C 始终保持白色 → 被当作垃圾回收 → 但 A 还在用 C → 悬挂指针！
```

如果漏标发生，白色对象被错误回收，导致悬挂指针——这是致命错误。

## 3. 写屏障（Write Barrier）

为了在并发标记中防止漏标，Go 使用**写屏障**机制——在指针赋值时插入额外的标记逻辑。

### 3.1 插入写屏障（Dijkstra Write Barrier）

思路：当黑色对象 A 新增对白色对象 C 的引用时，将 C 标记为灰色。

```
writePointer(slot, ptr):
    shade(ptr)        // 将 ptr 指向的对象标记为灰色
    *slot = ptr
```

**优点**：破坏漏标条件 1，保证黑色对象不会直接引用白色对象。

**缺点**：栈上的指针赋值不会触发写屏障（性能原因），因此标记结束时需要 **STW 重新扫描所有栈**。

### 3.2 删除写屏障（Yuasa Write Barrier）

思路：当灰色对象 B 删除对白色对象 C 的引用时，将 C 标记为灰色。

```
writePointer(slot, ptr):
    shade(*slot)      // 将旧的引用目标标记为灰色
    *slot = ptr
```

**优点**：破坏漏标条件 2，保证删除的引用目标不会丢失。

**缺点**：GC 开始时需要 STW 扫描所有栈建立快照；且会产生浮动垃圾（本应回收的对象被保留到下一轮 GC）。

### 3.3 混合写屏障（Hybrid Write Barrier，Go 1.8+）

Go 1.8 引入混合写屏障，结合插入和删除写屏障的优点：

```
writePointer(slot, ptr):
    shade(*slot)      // 旧引用目标标记为灰色（删除屏障）
    shade(ptr)        // 新引用目标标记为灰色（插入屏障）
    *slot = ptr
```

配合策略：
1. GC 开始时，将**所有栈上的对象**全部标记为黑色（无需 STW 扫描栈）
2. 堆上的指针赋值通过混合写屏障保护
3. 新创建的对象直接标记为黑色

**优点**：
- 消除了标记结束时 STW 重新扫描栈的需求
- STW 时间降至亚毫秒级

## 4. GC 完整流程

Go 的 GC 分为四个阶段：

```
 ┌──────────────────────────────────────────────────────┐
 │  阶段 1       阶段 2         阶段 3       阶段 4      │
 │ Mark Setup  Concurrent    Mark         Concurrent   │
 │  (STW)      Marking     Termination    Sweeping     │
 │             (并发)        (STW)         (并发)       │
 └──────────────────────────────────────────────────────┘
       ↑                        ↑
    极短 STW                  极短 STW
   (通常 < 30μs)            (通常 < 30μs)
```

### 阶段 1：Mark Setup（标记准备，STW）

- 开启写屏障
- 将所有 P 的 mcache 中的 tiny allocator 缓存 flush
- 将所有栈标记为"需要扫描"

**此阶段 STW 非常短**，通常 10-30 微秒。

### 阶段 2：Concurrent Marking（并发标记）

- GC goroutine 和 mutator 并行运行
- 从 GC Root 开始三色标记（扫描栈、全局变量、堆中的指针）
- 使用 **GC Worker**（后台标记 goroutine）执行标记工作
- 默认占用 **25%** 的 CPU 时间用于 GC（GOMAXPROCS=4 时 1 个 P 专职 GC）
- 写屏障保证并发安全

### 阶段 3：Mark Termination（标记终止，STW）

- 确保所有标记工作完成
- 关闭写屏障
- 执行清理工作

**此阶段 STW 也非常短**。

### 阶段 4：Concurrent Sweeping（并发清扫）

- 回收白色对象占用的 mspan
- 与 mutator 并发进行
- 清扫是惰性的：在分配内存时按需清扫对应的 mspan

## 5. GC 触发条件

| 触发方式 | 说明 |
|----------|------|
| 堆增长触发 | 堆内存增长到上次 GC 后存活大小的 `1 + GOGC/100` 倍时触发 |
| 定时触发 | `sysmon` 监控：超过 2 分钟未 GC 时强制触发 |
| 手动触发 | `runtime.GC()` 手动触发（阻塞直到 GC 完成） |

### GOGC 参数

`GOGC` 控制 GC 频率，默认值 **100**，表示当堆内存增长 100% 时触发 GC。

```
下次 GC 目标堆大小 = 上次 GC 后存活堆大小 × (1 + GOGC / 100)
```

| GOGC 值 | 效果 |
|---------|------|
| 100（默认） | 堆翻倍时触发 GC |
| 50 | 堆增长 50% 即触发（GC 更频繁，内存占用更低） |
| 200 | 堆增长 200% 才触发（GC 更少，内存占用更高） |
| `off` | 关闭 GC（不推荐） |

```bash
GOGC=200 ./myapp
```

或在代码中：

```go
import "runtime/debug"

debug.SetGCPercent(200)
```

### SetMemoryLimit（Go 1.19+）

`GOGC` 的问题：只关注增长比例，无法限制绝对内存用量。Go 1.19 引入 Soft Memory Limit：

```go
debug.SetMemoryLimit(1 << 30) // 限制 ~1GB
```

或环境变量：

```bash
GOMEMLIMIT=1GiB ./myapp
```

配合 `GOGC=off` 使用时：GC 仅在接近内存限制时触发，可减少不必要的 GC 开销。

```bash
GOGC=off GOMEMLIMIT=1GiB ./myapp
```

## 6. GC 调优

### 6.1 减少堆上的对象分配

这是最根本的优化——GC 压力来自堆上的对象数量和大小。

**方法 1：对象复用（sync.Pool）**

```go
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    // 使用 buf ...
}
```

`sync.Pool` 是 GC 友好的对象池：
- GC 时 Pool 中的对象可能被回收（非永久缓存）
- 适合生命周期短、创建开销大的临时对象（如 buffer、临时结构体）
- 标准库中 `fmt`、`encoding/json` 等大量使用

**方法 2：预分配（避免动态扩容）**

```go
// 差：多次扩容，产生多个被丢弃的底层数组
s := make([]int, 0)
for i := 0; i < 10000; i++ {
    s = append(s, i)
}

// 好：一次分配
s := make([]int, 0, 10000)
for i := 0; i < 10000; i++ {
    s = append(s, i)
}
```

**方法 3：值类型代替指针**

```go
// 差：每个 Point 都是堆上的独立对象
type Line struct {
    Start *Point
    End   *Point
}

// 好：Point 内联到 Line 中，减少堆对象数量
type Line struct {
    Start Point
    End   Point
}
```

**方法 4：字符串拼接使用 strings.Builder**

```go
// 差：每次 += 都产生新字符串
s := ""
for i := 0; i < 1000; i++ {
    s += "a"
}

// 好：strings.Builder 内部只分配一次
var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("a")
}
s := b.String()
```

### 6.2 减少指针类型

GC 需要扫描所有包含指针的对象。减少指针可以减少 GC 扫描量。

```go
// 差：map 的 key 和 value 都是指针类型，GC 需要扫描所有 entry
map[string]*Data

// 好（如果可以接受值拷贝）：value 为值类型，noscan span 不参与 GC 扫描
map[string]Data

// 差：切片中的每个元素都是指针
[]*Data

// 好：连续内存，一个 mspan，GC 扫描更快
[]Data
```

### 6.3 调整 GOGC 和 GOMEMLIMIT

不同场景下的推荐策略：

| 场景 | 推荐设置 | 说明 |
|------|----------|------|
| 内存充足、追求低延迟 | `GOGC=200` 或更高 | 减少 GC 频率 |
| 内存受限（容器） | `GOMEMLIMIT=XMiB` | 硬性内存上限 |
| 极致低 GC 开销 | `GOGC=off` + `GOMEMLIMIT` | 仅在接近限制时 GC |
| 内存敏感 | `GOGC=50` | 更频繁 GC，控制内存 |

### 6.4 使用 ballast（Go 1.19 之前的技巧）

Go 1.19 之前没有 `GOMEMLIMIT`，可通过分配一个大的空数组来稳定 GC 频率：

```go
func main() {
    // 分配 1GB 的 ballast（不含指针，不会被 GC 扫描）
    ballast := make([]byte, 1<<30)
    _ = ballast

    // 应用逻辑 ...
}
```

原理：ballast 使存活堆大小变大，GOGC 基于存活堆大小计算，因此 GC 触发的绝对阈值更高。

**Go 1.19+ 建议使用 `GOMEMLIMIT` 替代 ballast。**

### 6.5 Off-heap 存储

对于超大量数据（如缓存），可以使用堆外内存或外部存储，完全绕过 GC：

- **mmap**：直接映射文件到内存
- **cgo + malloc**：通过 C 分配堆外内存
- **外部缓存**：Redis、Memcached 等
- 开源库：`github.com/coocood/freecache`（零 GC 开销的 Go 缓存）、`github.com/dgraph-io/ristretto`

## 7. GC 可视化与诊断

### 7.1 gctrace

```bash
GODEBUG=gctrace=1 ./myapp
```

输出示例：

```
gc 1 @0.012s 2%: 0.021+1.2+0.009 ms clock, 0.17+0.8/1.0/0+0.073 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
```

各字段含义：

```
gc 1          — 第 1 次 GC
@0.012s       — 程序启动后 0.012 秒
2%            — GC 占用 CPU 时间百分比

0.021+1.2+0.009 ms clock:
  0.021       — STW Mark Setup 耗时
  1.2         — 并发标记耗时
  0.009       — STW Mark Termination 耗时

0.17+0.8/1.0/0+0.073 ms cpu:
  0.17        — STW Mark Setup CPU 时间
  0.8/1.0/0   — 标记阶段：辅助标记 / 后台标记 / 空闲标记
  0.073       — STW Mark Termination CPU 时间

4->4->1 MB:
  4           — GC 开始时堆大小
  4           — GC 结束时堆大小
  1           — GC 后存活对象大小

5 MB goal    — 下次 GC 的堆大小目标
8 P          — P 的数量
```

### 7.2 pprof

```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

```bash
# 查看堆内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看内存分配次数（allocs 视角）
go tool pprof -alloc_objects http://localhost:6060/debug/pprof/heap

# 可视化（需安装 graphviz）
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

常用 pprof 命令：

| 命令 | 说明 |
|------|------|
| `top` | 查看内存分配 Top N |
| `list funcName` | 查看某函数的逐行分配 |
| `web` | 浏览器中查看调用图 |
| `tree` | 树状显示调用关系 |

### 7.3 go tool trace

```go
import "runtime/trace"

f, _ := os.Create("trace.out")
trace.Start(f)
defer trace.Stop()
```

```bash
go tool trace trace.out
```

可查看：GC 活动时间线、goroutine 调度、STW 暂停时长等。

### 7.4 runtime.ReadMemStats

```go
var ms runtime.MemStats
runtime.ReadMemStats(&ms)

fmt.Printf("HeapAlloc    = %d MB\n", ms.HeapAlloc/1024/1024)
fmt.Printf("HeapInuse    = %d MB\n", ms.HeapInuse/1024/1024)
fmt.Printf("HeapIdle     = %d MB\n", ms.HeapIdle/1024/1024)
fmt.Printf("HeapReleased = %d MB\n", ms.HeapReleased/1024/1024)
fmt.Printf("NumGC        = %d\n", ms.NumGC)
fmt.Printf("PauseTotalNs = %d ms\n", ms.PauseTotalNs/1e6)
fmt.Printf("LastPause    = %d μs\n", ms.PauseNs[(ms.NumGC+255)%256]/1e3)
```

| 字段 | 含义 |
|------|------|
| `HeapAlloc` | 堆上已分配且正在使用的内存 |
| `HeapSys` | 从 OS 获取的堆内存总量 |
| `HeapIdle` | 空闲的 span 内存 |
| `HeapReleased` | 已归还给 OS 的内存 |
| `HeapObjects` | 堆上的对象数量（影响 GC 标记时间） |
| `NumGC` | GC 累计次数 |
| `PauseTotalNs` | 所有 GC STW 暂停总时间 |
| `GCCPUFraction` | GC 占用的 CPU 比例 |

## 8. GC 优化实战 Checklist

```
 1. [ ] 使用 pprof 定位堆内存分配热点（top + list）
 2. [ ] 开启 gctrace 观察 GC 频率和 STW 时间
 3. [ ] 检查 HeapObjects 数量，减少小对象分配
 4. [ ] 高频创建销毁的对象 → sync.Pool
 5. [ ] slice/map 预分配容量
 6. [ ] 值类型代替指针类型（减少 GC 扫描量）
 7. [ ] 字段重排优化内存对齐
 8. [ ] 大缓存考虑 off-heap 或外部存储
 9. [ ] 容器环境设置 GOMEMLIMIT
10. [ ] 调整 GOGC（低延迟 → 调高，低内存 → 调低）
```

## 参考资料

- [Go GC 源码](https://github.com/golang/go/blob/master/src/runtime/mgc.go)
- [A Guide to the Go Garbage Collector](https://tip.golang.org/doc/gc-guide)
- [Getting to Go: The Journey of Go's Garbage Collector](https://go.dev/blog/ismmkeynote)
- [Go 1.19 Soft Memory Limit](https://tip.golang.org/doc/gc-guide#Memory_limit)
- [Garbage Collection In Go (William Kennedy)](https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html)
