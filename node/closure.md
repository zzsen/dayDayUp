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

## 闭包常见应用场景
1. 作为返回值
    ``` javascript
    function returnFunc() {
        var a = 1
        return function innerFunc() {
            console.log(a)
        }
    }
    var outterFunc = returnFunc()
    outterFunc()  // 1
    ```

2. 作为参数传递
    ``` javascript
    function foo() {
        var a = 1
        function argFunc() {
            console.log(a)
        }
        emitFunc(argFunc)
    }

    function emitFunc(argFunc) {
        argFunc()
    }
    emitFunc(foo) // 1
    ```

3. 封装私有变量
    ``` javascript
    function createCounter() {
        let count = 0
        
        return {
            increment: function() {
                count++
            },
            decrement: function() {
                count--
            },
            getCount: function() {
                return count
            }
        }
    }

    const counter = createCounter()
    counter.increment()
    console.log(counter.getCount()) // 输出: 1
    ```

4. 模块化开发

    用闭包模拟私有方法
    ``` javascript
    var Counter = (function() {
    var privateCounter = 0
    function changeBy(val) {
        privateCounter += val
    }
    return {
        increment: function() {
            changeBy(1)
        },
        decrement: function() {
            changeBy(-1)
        },
        value: function() {
            return privateCounter
        }
    }
    })()

    console.log(Counter.value()) /* logs 0 */
    Counter.increment()
    Counter.increment()
    console.log(Counter.value()) /* logs 2 */
    Counter.decrement()
    console.log(Counter.value()) /* logs 1 */
    ```

5. 面向时间编程

    定时器、事件监听、Ajax 请求、跨窗口通信、Web Workers 或者任何异步，只要使用了回调函数，实际上就是在使用闭包
    ```javascript
    // 定时器
    function wait(message) {
      setTimeout( function timer() {
        console.log( message )
      }, 1000)
    }
    wait( "Hello, closure!" )
    // message 是 wait 函数的变量，但是被 timer 函数引用，就形成了闭包
    // 调用 wait 后，wait 函数压入调用栈，message 被赋值，并调用定时器任务，随后弹出，1000ms之后，回调函数timer 压入调用栈，因为引用 message，所以就能打印出 message

    // 事件监听 
    let a = 1
    let btn = document.getElementById('btn')
    btn.addEventListener('click', function callback() {
      console.log(a)
    })
    // 变量 a 被 callback 函数引用，形成闭包
    // 事件监听和定时器一样，都属于把函数作为参数传递形成的闭包。addEventListener函数有两个参数，一为事件名，二为回调函数
    // 调用事件监听函数，将 addEventListener 压入调用栈，词法环境中有 click 和 callback 等变量，并因为 callback 为函数，并有作用域函数形成，引用 a 变量。之后弹出调用栈，当用户点击时，回调函数触发，callback 函数压入调用栈，a 沿着作用域链往上找，找到全局作用域中的变量 a，并打印出

    // AJAX
    let a = 1
    fetch("/api").then(function callback() {
      console.log(a)
    })
    // 同事件监听
    ```

## 注意事项
闭包使用不当，容易造成内存泄漏或性能问题
### 内存泄漏
由于闭包会保留对外部函数作用域的引用，如果没有正确释放，会导致内存泄漏
``` javascript
function outerFunction() {
  var data = 'data'
  return function innerFunction() {
  console.log(data)
  }
}
// 这里leakedFunction保留了对outerFunction作用域的引用
var leakedFunction = outerFunction()
// leakedFunction仍然保留了对outerFunction中的data的引用，导致data无法被垃圾回收，从而导致内存泄漏

// 解决方案之一：手动释放引用
leakedFunction = null //当不再需要leakedFunction时，需要手动解除引用
```

### 性能问题
``` javascript
function outerFunction() {
  var data = 'data'
  return function innerFunction() {
    console.log(data)
  }
}
for (var i = 0; i < 10000; i++) {
  var fn = outerFunction()
  // 在每次循环中，都会创建一个新的闭包函数
  // 执行一些操作...
  fn = null
  // 但是没有手动解除对闭包函数的引用
}

// 循环中，创建了10000个闭包函数，每个函数都保留了对outerFunction作用域的引用，会占用大量的内存，无法被垃圾回收，从而导致性能问题
```

### 参考解决方案
* 在不需要使用闭包函数时，手动解除对其的引用，如，将其赋值为null
* 尽量避免在循环中创建大量闭包，可以将闭包移出循环，或者使用其他方式实现相同功能
* 注意闭包函数中对外部变量的作用，确保不会保留对不在需要的变量的引用

## 相关文档
[深入理解JavaScript——闭包](https://zhuanlan.zhihu.com/p/574913236)
