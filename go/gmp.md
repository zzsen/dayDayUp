# GMP 调度模型

## 1. 概述

GMP 是 Go 运行时（runtime）调度器的核心模型，负责将 goroutine 调度到操作系统线程上执行。

| 缩写 | 全称 | 说明 |
|------|------|------|
| **G** | Goroutine | 协程，Go 中的并发执行单元，包含栈、指令指针等信息 |
| **M** | Machine | 工作线程，对应一个操作系统线程（OS Thread） |
| **P** | Processor | 逻辑处理器，管理本地 G 队列，是 G 和 M 之间的桥梁 |

三者关系：**G 需要绑定到 P 上，P 需要绑定到 M 上，才能被执行**。

```
              ┌───────────────────────────────────┐
              │           Go Scheduler            │
              │                                   │
              │  ┌─────────── 全局队列 ───────────┐ │
              │  │  G   G   G   G   G   ...      │ │
              │  └───────────────────────────────┘ │
              │                                   │
              │   ┌─────┐      ┌─────┐            │
              │   │  P  │      │  P  │   ...      │
              │   │本地队│      │本地队│            │
              │   │G G G│      │G G G│            │
              │   └──┬──┘      └──┬──┘            │
              │      │            │               │
              │   ┌──▼──┐      ┌──▼──┐            │
              │   │  M  │      │  M  │   ...      │
              │   └──┬──┘      └──┬──┘            │
              └──────┼────────────┼───────────────┘
                     │            │
              ┌──────▼────────────▼───────────────┐
              │         操作系统内核                │
              │     OS Thread    OS Thread         │
              └───────────────────────────────────┘
```

## 2. 为什么需要 GMP 模型

### 2.1 早期方案（GM 模型）的问题

Go 1.1 之前没有 P，调度器使用 GM 模型：所有 M 从**全局队列**中获取 G 来执行。

存在的问题：

1. **全局队列锁竞争**：所有 M 从同一个全局队列获取 G，需要加锁，高并发场景下锁竞争严重
2. **M 转移 G 开销大**：当 M1 上的 G1 创建了 G2，G2 会被放入全局队列，而 G2 大概率会被 M2 执行，导致局部性（locality）差
3. **M 的内存缓存（mcache）浪费**：每个 M 都有 mcache（内存分配缓存），但阻塞中的 M 的 mcache 完全闲置

### 2.2 引入 P 的改进

Go 1.1 引入 P（Processor）后：

1. **本地队列无锁化**：每个 P 有自己的本地队列（Local Run Queue），P 上的 M 优先从本地队列取 G，无需加锁
2. **提高局部性**：M1 创建的 G2 优先放入 M1 所绑定的 P 的本地队列，大概率在同一个 M 上执行
3. **mcache 绑定 P 而非 M**：mcache 跟着 P 走，阻塞的 M 释放 P 后，P 带着 mcache 绑定到新的 M 上，避免浪费

## 3. G（Goroutine）

G 是 Go 中用户级的轻量级线程，由 Go 运行时管理，而非操作系统。

### 3.1 核心特点

- 初始栈大小仅 **2KB**（OS 线程通常 1-8MB），可动态增长
- 创建和销毁的开销远小于 OS 线程
- 通过 `go` 关键字创建

### 3.2 G 的状态

```
                 ┌─────────┐
          ┌─────►│ _Gwaiting│◄──────┐
          │      └────┬────┘       │
          │           │            │
    chan/IO/锁     被唤醒       chan/IO/锁
          │           │            │
          │      ┌────▼────┐       │
  ┌───────┴─────►│_Grunnable│──────┤
  │              └────┬────┘       │
  │                   │            │
  │              被调度执行          │
  │                   │            │
  │              ┌────▼────┐       │
  │              │_Grunning │───────┘
  │              └────┬────┘
  │                   │
  │              执行完毕
  │                   │
  │              ┌────▼────┐
  └──────────────│  _Gdead  │
                 └─────────┘
```

| 状态 | 说明 |
|------|------|
| `_Gidle` | 刚分配，尚未初始化 |
| `_Grunnable` | 在运行队列中，等待被调度执行，尚未拥有栈 |
| `_Grunning` | 正在 M 上执行，拥有栈，已分配 M 和 P |
| `_Gsyscall` | 正在执行系统调用，拥有栈，未分配 P（M 上有但不计入 P） |
| `_Gwaiting` | 被阻塞（channel、锁、IO、GC 等），不在运行队列中 |
| `_Gdead` | 未被使用（可能刚退出或刚初始化），可被复用 |

### 3.3 关键源码结构

```go
// src/runtime/runtime2.go

type g struct {
    stack       stack   // 栈内存范围 [stack.lo, stack.hi)
    stackguard0 uintptr // 栈溢出检测（用于抢占调度）
    m           *m      // 当前绑定的 M
    sched       gobuf   // 调度上下文（SP、PC、BP 等寄存器）
    atomicstatus atomic.Uint32 // G 的状态
    goid         uint64 // goroutine ID
    preempt      bool   // 抢占标记
    // ...
}

type gobuf struct {
    sp   uintptr // 栈指针
    pc   uintptr // 程序计数器
    g    guintptr
    ret  uintptr
    bp   uintptr // 基址指针（用于帧指针）
    // ...
}
```

`gobuf` 保存了 G 的调度上下文，调度器切换 G 时只需保存和恢复 `gobuf` 中的寄存器信息。

## 4. M（Machine / OS Thread）

M 代表一个操作系统线程，由操作系统管理和调度。

### 4.1 核心特点

- Go 程序启动时默认最多创建 **10000** 个 M（`runtime/debug.SetMaxThreads` 可调整）
- M 必须绑定一个 P 才能执行 G（系统调用等特殊情况除外）
- 当 M 因系统调用阻塞时，会释放 P 给其他 M 使用

### 4.2 关键源码结构

```go
// src/runtime/runtime2.go

type m struct {
    g0      *g     // 调度栈，用于执行调度代码
    curg    *g     // 当前正在运行的 G
    p       puintptr // 当前绑定的 P
    nextp   puintptr // 唤醒时优先绑定的 P
    oldp    puintptr // 系统调用前绑定的 P
    spinning bool   // 是否处于自旋状态（正在找可运行的 G）
    // ...
}
```

每个 M 都有一个特殊的 **g0**，g0 使用操作系统分配的栈（通常 8MB），用于执行调度逻辑（如 `schedule()`、`findRunnable()` 等）。普通 G 的调度切换流程为：`G → g0（执行调度） → G'`。

### 4.3 M0

M0 是程序启动后的第一个 M（主线程），负责：
- 执行初始化操作（`runtime.init`）
- 启动第一个 G（即 `main.main`）

M0 保存在全局变量 `runtime.m0` 中，不需要在堆上分配。

## 5. P（Processor）

P 是逻辑处理器，是 G 和 M 之间的桥梁。

### 5.1 核心特点

- P 的数量由 `GOMAXPROCS` 决定（默认等于 CPU 核数）
- P 持有本地 G 队列（最多 256 个）
- P 持有 mcache（内存分配缓存），避免每次内存分配都需要全局锁

### 5.2 P 的状态

| 状态 | 说明 |
|------|------|
| `_Pidle` | 空闲，未与 M 绑定，在空闲 P 列表中 |
| `_Prunning` | 正在被 M 使用，此时 M 正在执行 G 或调度代码 |
| `_Psyscall` | M 进入系统调用，P 可能被其他 M 抢走 |
| `_Pgcstop` | GC 导致的 STW（Stop The World），暂停 |
| `_Pdead` | P 的数量被缩减（如动态调整 GOMAXPROCS）后多余的 P |

### 5.3 关键源码结构

```go
// src/runtime/runtime2.go

type p struct {
    id          int32
    status      uint32   // P 的状态
    m           muintptr // 绑定的 M
    mcache      *mcache  // 内存分配缓存

    // 本地可运行 G 队列（无锁环形队列）
    runqhead uint32
    runqtail uint32
    runq     [256]guintptr

    runnext guintptr // 下一个优先执行的 G（优先级最高）
    // ...
}
```

`runnext` 字段很关键：当前 G 通过 `go` 创建的新 G 会优先放入 `runnext`，下次调度时优先执行，提高了局部性。

## 6. 调度流程

### 6.1 G 的创建与入队

当执行 `go func()` 时：

1. 调用 `runtime.newproc` 创建新的 G
2. 优先复用 `_Gdead` 状态的空闲 G（从 P 的 `gFree` 列表或全局 `sched.gFree`）
3. 将新 G 放入当前 P 的 `runnext`（最高优先级），被挤出的旧 `runnext` 放入本地队列尾部
4. 如果本地队列已满（256 个），将本地队列的**前半部分**连同新 G 一起放入全局队列

### 6.2 G 的调度执行

M 调用 `schedule()` 获取可运行的 G，查找顺序：

```
1. 当前 P 的 runnext       ← 最高优先级
      ↓ 空
2. 当前 P 的本地队列        ← 无锁访问
      ↓ 空
3. 全局队列                ← 需要加锁，一次取多个（min(全局队列长度/GOMAXPROCS+1, 本地队列容量/2)）
      ↓ 空
4. 网络轮询器（netpoll）    ← 检查就绪的网络 IO
      ↓ 空
5. 从其他 P 偷（steal）     ← work stealing，偷一半
```

### 6.3 调度触发时机

| 触发方式 | 说明 |
|----------|------|
| 主动让出 | `runtime.Gosched()` 主动让出 CPU |
| 系统调用 | 进入系统调用（`_Gsyscall`），M 释放 P |
| 阻塞操作 | channel、mutex、网络 IO 等阻塞时进入 `_Gwaiting` |
| 抢占调度 | 运行超过 10ms 被信号抢占（Go 1.14+ 基于信号的异步抢占） |
| 栈增长检查 | 函数调用时的栈溢出检查点同时检查抢占标记 |

## 7. 调度策略

### 7.1 Work Stealing（工作窃取）

当一个 P 的本地队列为空时，会尝试从其他 P 的本地队列中**偷取一半的 G** 来执行。

```
    P1（空闲）               P2（繁忙）
    ┌────────┐              ┌────────┐
    │ 本地队列│              │ 本地队列│
    │  (空)   │   ← 偷一半 ─ │ G G G G│
    └────────┘              └────────┘
```

这样可以避免某些 P 饿死（idle），提高整体 CPU 利用率。

### 7.2 Hand Off（交接机制）

当 M 因为系统调用或其他原因阻塞时，会将 P 交给（hand off）其他空闲的 M 使用，避免 P 及其本地队列中的 G 被阻塞。

```
  阻塞前:               阻塞后:
  ┌───┐                 ┌───┐
  │ M1│ ← 阻塞           │ M1│ (阻塞中，无 P)
  │ P │                  └───┘
  │G G│
  └───┘                 ┌───┐
                        │ M2│ ← 接管
                        │ P │
                        │G G│
                        └───┘
```

流程：
1. M1 进入系统调用，运行时检测到 P 上仍有 G 待执行
2. M1 释放 P
3. 运行时唤醒或创建 M2，将 P 绑定到 M2
4. M1 系统调用返回后，尝试获取空闲 P；若无空闲 P，M1 上的 G 放入全局队列，M1 进入休眠

### 7.3 抢占式调度

#### Go 1.13 及之前：协作式抢占

依赖函数调用时的栈增长检查点（`morestack`）来检测抢占标记。

问题：如果一个 G 执行了一个**没有函数调用的死循环**（如 `for {}`），则永远不会触发检查点，导致该 G 无法被抢占，其他 G 饿死。

```go
// 在 Go 1.13 及之前，以下代码会导致程序卡死
go func() {
    for {
        // 空循环，无函数调用，无法被抢占
    }
}()
```

#### Go 1.14+：基于信号的异步抢占

运行时启动了一个 `sysmon`（系统监控）goroutine，它会：
1. 检测运行超过 **10ms** 的 G
2. 向该 G 所在的 M 发送 `SIGURG` 信号
3. M 的信号处理函数中断当前 G 的执行，保存上下文，将 G 重新放入运行队列

这种方式不依赖函数调用，可以抢占任何正在执行的 G。

## 8. sysmon（系统监控）

`sysmon` 是运行时中的一个**监控函数**，运行在一个不绑定 P 的特殊 M 上，以独立的系统线程持续运行，负责：

| 职责 | 说明 |
|------|------|
| 抢占调度 | 检测运行超过 10ms 的 G，发送信号抢占 |
| 网络轮询 | 定期执行 `netpoll`，将就绪的 G 放入全局队列 |
| 回收 P | 将系统调用中超时的 P 回收（hand off） |
| 强制 GC | 超过 2 分钟未 GC 时强制触发 |

## 9. 完整调度生命周期示例

以下是一个 G 从创建到执行完毕的完整流程：

```
1. go func() 创建 G
      │
      ▼
2. G 放入当前 P 的 runnext 或本地队列
      │
      ▼
3. M 调用 schedule() 从 P 获取 G
      │
      ▼
4. M 调用 execute(G) 开始运行
      │
      ├──── 正常完成 ──► G 状态变为 _Gdead，放入空闲列表复用
      │
      ├──── channel 阻塞 ──► G 变为 _Gwaiting，M 切换到其他 G
      │                       被唤醒后 G 变为 _Grunnable 放回队列
      │
      ├──── 系统调用 ──► G 变为 _Gsyscall，P 被 hand off 给其他 M
      │                  系统调用返回后 G 变为 _Grunnable
      │
      └──── 运行超时 ──► sysmon 发送 SIGURG 信号抢占
                         G 变为 _Grunnable 放回队列
```

## 10. 关键参数

| 参数 | 默认值 | 说明 | 调整方式 |
|------|--------|------|----------|
| GOMAXPROCS | CPU 核数 | P 的数量，决定最大并行度 | `runtime.GOMAXPROCS(n)` 或环境变量 |
| MaxThreads | 10000 | M 的最大数量 | `runtime/debug.SetMaxThreads(n)` |
| 本地队列容量 | 256 | 每个 P 的本地队列最大 G 数 | 不可调整（编译时常量） |
| 抢占时间阈值 | 10ms | G 运行超过该时间会被抢占 | 不可调整 |

## 11. 常见问题

### Q1: GOMAXPROCS 设多少合适？

- **CPU 密集型**：设为 CPU 核数（默认值），过多反而增加上下文切换开销
- **IO 密集型**：可适当大于 CPU 核数，因为大部分 G 在等待 IO，不占用 CPU
- 容器环境下注意：Go 默认读取的是**宿主机**的 CPU 核数，建议使用 `go.uber.org/automaxprocs` 自动适配容器 CPU 限制

### Q2: goroutine 泄漏怎么排查？

goroutine 泄漏通常因为 goroutine 被阻塞无法退出（如读取没有写入者的 channel）。排查手段：

- `runtime.NumGoroutine()` 监控 goroutine 数量
- `pprof`（`net/http/pprof` 或 `runtime/pprof`）查看 goroutine 堆栈
- `go vet` / `goleak`（`go.uber.org/goleak`）在测试中检测泄漏

### Q3: 为什么 goroutine 比线程轻量？

| 对比项 | goroutine | OS Thread |
|--------|-----------|-----------|
| 初始栈大小 | 2KB（可动态增长） | 1-8MB（固定） |
| 创建开销 | 用户态，约 0.3μs | 内核态，约 10-30μs |
| 上下文切换 | 用户态，仅保存少量寄存器 | 内核态，保存全部寄存器 + 内核栈切换 |
| 调度 | Go runtime（用户态） | 操作系统内核 |
| 数量上限 | 轻松支持数十万 | 通常数千即有压力 |

## 参考资料

- [Go 调度器源码](https://github.com/golang/go/blob/master/src/runtime/proc.go)
- [Scalable Go Scheduler Design Doc (by Dmitry Vyukov)](https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw)
- [GMP 模型详解 (Go 夜读)](https://talkgo.org/t/topic/29)
- [Go: Goroutine, OS Thread and CPU Management](https://medium.com/a-journey-with-go/go-goroutine-os-thread-and-cpu-management-2f5a5eaf518a)
