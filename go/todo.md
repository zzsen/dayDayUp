## 前言
该文档主要记录学习go过程中遇到的一些后续需要搞明白的问题

## 问题
1.gorm文档中说到模型实现了 `Scanner`和`Value`接口, 什么是 [Scanner](https://pkg.go.dev/database/sql#Scanner) 和 [Valuer](https://pkg.go.dev/database/sql/driver#Valuer)?

>[gorm-模型定义](https://gorm.io/zh_CN/docs/models.html)
>
>`模型是标准的 struct，由 Go 的基本数据类型、实现了 Scanner 和 Valuer 接口的自定义类型及其指针或别名组成`
