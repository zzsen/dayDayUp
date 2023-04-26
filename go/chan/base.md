# 通道

## 1. 概念

通道是 goroutine 之间的通道, 可以让 goroutine 之间相互通讯. 通道数据类型是通道允许传输的数据类型。通道的零值为 nil。nil 通道没有任何用处，因此通道必须使用类似于 map 和切片的方法来定义。

## 2. 声明

```go
/// 方式1
// var [通道名] chan [通道数据类型]
var c chan int
// [通道名] = make(chan [通道数据类型])
c = make(chan int)

/// 方式2
// [通道名] := make(chan [通道数据类型])
ch := make(chan int)
```

## 3. 通道是引用类型

通道是引用类型, 作为参数传递时, 传递的是内存地址

```go
package main

import (
	"fmt"
)

func main() {
	ch := make(chan int)
	fmt.Printf("类型: %T, 地址: %p\n", ch, ch) // 类型: chan int, 地址: 0xc00001e0c0
	testFunc(ch)
}

func testFunc(ch chan int) {
	fmt.Printf("类型: %T, 地址: %p\n", ch, ch) // 类型: chan int, 地址: 0xc00001e0c0
}
```

## 4. 使用

1. chan 是同步的, **同一时间, 只能有一个 goroutine 操作**

2. 阻塞
   通道的发送和接受默认都是会发生阻塞，所以**通道的发送和接收必须处在不同的 goroutine 中**
   发送数据：`chan <- data`, 阻塞的，直到另一条 goroutine，读取数据来解除阻塞
   读取数据：`data <- chan`, 阻塞的。直到另一条 goroutine，写出数据解除阻塞。

3. 通道读写
   `chan <- data`, 发送数据到通道, 向通道中写数据
   `data <- chan`, 从通道中获取数据, 从通道中读数据

### 4.1 阻塞

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	dataChan := make(chan int)
	done := make(chan bool) // 通道

	go func() {
		fmt.Println("子goroutine执行中")
		time.Sleep(3 * time.Second)
		data := <-dataChan // 从通道中读取数据
		fmt.Println("接受到data：", data)
		done <- true
	}()

	// 向通道中写数据。。
	time.Sleep(1 * time.Second)
	dataChan <- 117

	<-done
	fmt.Println("执行完成")
}


// 子goroutine
// 接受到data： 117
// 执行完成
```

### 4.2 死锁

```go
package main

func main() {
	ch := make(chan int)
	ch <- 117
}

// fatal error: all goroutines are asleep - deadlock!

// goroutine 1 [chan send]:
// main.main()
//         D:/projects/dayDayUp/go/main.go:5 +0x31
// exit status 2
```

## 5. 关闭通道

```go
// close([通道名])
close(c)
```

接收通道数据时, 可判断通道是否已关闭

```go
// v: 通道中获取到的值,
// ok: 为true表示通道未关闭, 反之, 通道已关闭
v, ok := <- ch
```

**当通道已经关闭时, 读取到的数据是该类型的默认值**, 如 int 是 0, bool 是 false, string 是空字符串("")

```go
package main

import "fmt"

func main() {

	stringChan := make(chan string)
	close(stringChan)
	v1, ok := <-stringChan
	fmt.Printf("value: %v, ok: %v\n", v1, ok)

	boolChan := make(chan bool)
	close(boolChan)
	v2, ok := <-boolChan
	fmt.Printf("value: %v, ok: %v\n", v2, ok)

	intChan := make(chan int)
	close(intChan)
	v3, ok := <-boolChan
	fmt.Printf("value: %v, ok: %v\n", v3, ok)
}

// value: , ok: false
// value: false, ok: false
// value: false, ok: false
```

## 6.缓冲通道和非缓冲通道

非缓冲通道: 发送和接收都是阻塞的
缓冲通道: 发送数据到通道时, 只有通道缓冲区满时, 才会被阻塞; 读取数据时, 只有缓冲区为空时, 才会被阻塞

### 6.1 缓冲通道定义

```go
// [通道名] := make(chan [通道数据类型], [缓冲区容量])
c := make(chan int, 10)
```

## 7. time 包

time 包中有使用通道的实现定时器的 timer

```go
t:= time.NewTimer(d)
t:= time.AfterFunc(d, f)
c:= time.After(d)
```

### 7.1 time.NewTimer

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	//新建一个timer
	timer := time.NewTimer(5 * time.Second)
	fmt.Println(time.Now()) // 2023-04-25 11:30:46.1945963 +0800 CST m=+0.004986701

	//timer的channel中的信号，执行此段代码时会阻塞, 直到timer往chan中发送值
	ch2 := timer.C
	fmt.Println(<-ch2) // 2023-04-25 11:30:51.1966895 +0800 CST m=+5.007079901
}
```

查看`time.NewTimer`的源码, 如下:

```go
// NewTimer creates a new Timer that will send
// the current time on its channel after at least duration d.
func NewTimer(d Duration) *Timer {
	c := make(chan Time, 1)
	t := &Timer{
		C: c,
		r: runtimeTimer{
			when: when(d),
			f:    sendTime,
			arg:  c,
		},
	}
	startTimer(&t.r)
	return t
}
```

可以看出, `Timer`类型, 执行`NewTimer`时, 会先创建一个存放`Time`类型的`chan`, 创建个`Timer`, `startTimer`后, 返回 timer.
`runtimeTimer`源码如下:

```go
// Interface to timers implemented in package runtime.
// Must be in sync with ../runtime/time.go:/^type timer
type runtimeTimer struct {
	pp       uintptr
	when     int64
	period   int64
	f        func(any, uintptr) // NOTE: must not be closure
	arg      any
	seq      uintptr
	nextwhen int64
	status   uint32
}
```

计时器可在触发前, 主动 stop

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	// 计时器时长
	timerDuration := 1
	// 停止计时器时间
	timeToStopDuration := 3

	timer := time.NewTimer(time.Duration(timerDuration) * time.Second)

	go func() {
		//等触发时的信号
		<-timer.C
		fmt.Println("Timer计时结束")

	}()

	time.Sleep(time.Duration(timeToStopDuration) * time.Second)
	stop := timer.Stop()
	if stop {
		fmt.Println("Timer计时停止")
	}
}
// timerDuration < timeToStopDuration时, 输出: Timer计时结束
// timerDuration > timeToStopDuration时, 输出: Timer计时停止
```

### 7.2 time.After

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := time.After(3 * time.Second) // 3s后
	fmt.Println(time.Now())           // 2023-04-25 11:54:44.0327497 +0800 CST m=+0.004794601
	fmt.Println(<-ch)                 // 2023-04-25 11:54:47.0381495 +0800 CST m=+3.010194401
}
```

查看`time.After`源码, 如下

```go
// After waits for the duration to elapse and then sends the current time
// on the returned channel.
// It is equivalent to NewTimer(d).C.
// The underlying Timer is not recovered by the garbage collector
// until the timer fires. If efficiency is a concern, use NewTimer
// instead and call Timer.Stop if the timer is no longer needed.
func After(d Duration) <-chan Time {
	return NewTimer(d).C
}
```

本质是`NewTimer`的 chan
