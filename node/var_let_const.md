# 浅谈var let const
## var
在ES5中，顶层对象的属性和全局变量是等价的，用var声明的变量既是全局变量，也是顶层变量
``` javascript
var a = 10;
console.log(window.a) // 10
```
使用var声明的变量存在变量提升的情况
``` javascript
console.log(a) // undefined
var a = 20

// 编译阶段，编译器会将其变成以下执行
var a
console.log(a)
a = 20
```
使用var，可对一个变量进行多次声明，后面声明的变量会覆盖前面的变量声明
``` javascript
var a = 20 
var a = 30
console.log(a) // 30
```
在函数中使用使用var声明变量时候，该变量是局部的，否则是全局的
``` javascript
// 局部
var a = 20
function changeA(){
    var a = 30
}
changeA()
console.log(a) // 20 

// 全局
var b = 20
function changeB(){
   b = 30
}
changeB()
console.log(b) // 30 
```

## let
let是ES6新增的命令，用来声明变量

用法类似var，但所声明的变量，只在let命令所在的代码块内有效
``` javascript
{
    let a = 20
}
console.log(a) // ReferenceError: a is not defined.
```
不存在变量提升
``` javascript
console.log(a) // 报错ReferenceError
let a = 1
```
这表示在声明它之前，变量a是不存在的，这时如果用到它，会抛异常

只要**块级作用域**内存在let命令，这个区域就不再受外部影响
``` javascript
var a = 123
if (true) {
    a = 'abc' // ReferenceError
    let a;
}
```
使用let声明变量前，该变量都不可用，也就是常说的“**暂时性死区**”

let不允许在**相同作用域**中重复声明
``` javascript
// 相同作用域
let a = 20
let a = 30
// Uncaught SyntaxError: Identifier 'a' has already been declared


// 不同作用域
let b = 20
{
    let b = 30
}
```
因此，不能在函数内部重新声明参数
``` javascript
function func(arg) {
  let arg;
}
func()
// Uncaught SyntaxError: Identifier 'arg' has already been declared
```

## const
const用于声明只读常量，一旦声明，常量的值就不能改变
``` javascript
const a = 1
a = 3
// TypeError: Assignment to constant variable.
```
因此，const 在声明常量的同时必须初始化
``` javascript
const a;
// SyntaxError: Missing initializer in const declaration
```
禁止对用var或let声明过变量，再用const声明，会报错

**const实际上保证的并不是变量的值不得改动，而是变量指向的那个内存地址所保存的数据不得改动***
* 对于**简单类型**的数据，值就保存在变量指向的那个内存地址，因此等同于常量
* 对于**复杂类型**的数据，变量指向的内存地址，保存的只是一个指向实际数据的指针，const只能保证这个指针是固定的，并不能确保改变量的结构不变
``` javascript
const foo = {};

// 为 foo 添加一个属性，可以成功
foo.prop = 123;
foo.prop // 123

// 将 foo 指向另一个对象，就会报错
foo = {}; // TypeError: "foo" is read-only
```

## 区别
|特点|var|let|const|
|--|--|--|--|
|变量提升|Y|N|N|
|是否存在**暂时性死区**|N|Y|Y|
|是否存在**块级作用域**|N|Y|Y|
|是否允许重复声明|Y|N|N|
|是否允许修改声明的变量|Y|Y|N|


# 参考文献
[面试官：说说var、let、const之间的区别](https://vue3js.cn/interview/es6/var_let_const.html#%E5%9B%9B%E3%80%81%E5%8C%BA%E5%88%AB)
