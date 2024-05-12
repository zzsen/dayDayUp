# 闭包
## 基本概念
### 变量作用域
js有两种作用域：**全局作用域**和**函数作用域**，对应的变量分别称为**全局变量**和**局部变量**

``` javascript
// 全局变量
let globalVariavle = 1
function fn1() {
    console.log(globalVariavle)
}
fn1() // 1

// 局部变量
function fn2() {
    let localVariavle = 1
}
fn2()
console.log(localVariavle) // Uncaught ReferenceError: localVariavle is not defined
```
## 什么是闭包
可以访问其他函数内部定义的变量的函数，就是闭包。即在一个函数内部创建另一个函数，并且在这个内部函数可以访问外部函数的变量、参数和内部函数本身的局部变量。
``` javascript
function outterFunction() {
    let outVariavle = 1
    function innerFunction () {
        outVariable++
        console.log(outVariable)
    }
    return innerFunction
}
const closureFunction = outterFunction()
closureFunction() // 1
closureFunction() // 2
closureFunction() // 3
closureFunction() // 4

```
### 工作原理
当一个函数被定义时，它会创建一个作用域链(scope chain)，用于保存在函数内部定义的变量和函数。作用域链是一个由一系列变量对象(variable objects)组成的列表，每个变量对象对应一个包含变量和函数的作用域。当函数被执行时，会创建一个执行环境(execution context)，包含了函数的参数、局部变量和对应的作用域链。

当内部函数被定义时，它会创建一个闭包，并包含对其父函数作用域链的引用。这意味着内部函数可以访问父函数的变量和函数，以及父函数作用域链上的其他作用域。当内部函数被返回并在外部环境中被调用时，它仍然可以访问和操作这些变量。

## 闭包的特点
1. 函数嵌套函数
2. 内部函数可访问外部函数的变量
3. 被访问的参数和变量不会被JavaScript的垃圾回收机制回收

## 闭包的优缺点
### 优点
1. 保护函数内的变量安全，有利于代码封装
2. 延伸函数的作用范围，能够读取函数内部的变量
3. 在内存中维持一个变量
    
    不会被垃圾回收机制销毁，但用得过多就变成缺点了，占内存
    > 没啥用的哲学小知识：矛盾双方在一定条件下相互转化

### 缺点
1. 常驻内存，增加性能开销，使用不当会造成**内存泄漏**



## 相关文档
[深入理解JavaScript——闭包](https://zhuanlan.zhihu.com/p/574913236)
