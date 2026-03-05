# Go 学习笔记

## 运行时原理

| 主题 | 说明 |
|------|------|
| [GMP 调度模型](./gmp.md) | G/M/P 概念、调度流程、Work Stealing、Hand Off、抢占式调度 |
| [内存管理](./memory.md) | TCMalloc 架构、mcache/mcentral/mheap、三类对象分配策略、内存对齐 |
| [垃圾回收（GC）](./gc.md) | 三色标记、写屏障、GC 四阶段、GOGC/GOMEMLIMIT 调优、诊断工具 |
| [内存逃逸](./escape_to_heap.md) | 逃逸分析场景（指针返回、动态类型、闭包、栈空间不足等） |

## 并发编程

| 主题 | 说明 |
|------|------|
| [Channel 基础](./chan/base.md) | 通道概念、缓冲/非缓冲通道、select、Timer |
| [读写已关闭的 Channel](./chan/write_or_read_a_closed_chan.md) | 对已关闭 chan 读写的行为与原因 |
| [for-select 与已关闭的 Channel](./chan/for_select_handle_a_closed_chan.md) | for-select 中 channel 关闭后的处理方式 |

## 工程实践

| 主题 | 说明 |
|------|------|
| [静态代码分析工具](./lint.md) | vet、golangci-lint 等工具的使用 |

## Gin 框架

- [gin 快速入门指引](https://blog.csdn.net/zzsan/article/details/120532857)
- [浅谈 gin](https://blog.csdn.net/zzsan/article/details/120458301)
- [go 异常处理 & 错误堆栈获取](https://blog.csdn.net/zzsan/article/details/123521653)

## 基础入门

- [go 开发前期准备](https://blog.csdn.net/zzsan/article/details/120356705)
- [go 指令和 mod 文件解析](https://blog.csdn.net/zzsan/article/details/120375203)

## 外部参考

- [Go 为什么这么"快" — 腾讯技术工程（知乎）](https://zhuanlan.zhihu.com/p/111346689)
