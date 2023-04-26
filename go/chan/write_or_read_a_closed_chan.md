# 对已经关闭的的chan进行读写，会怎么样？为什么？
## 答案
* 写已经关闭的 chan 会 panic
* 读已经关闭的 chan 能一直读到内容，但是读到的内容和**通道内关闭前通道内有无元素**有关。
    * 如果 chan 关闭前，通道内有元素还未读 , 会正确读到 chan 内的值，且返回的第二个 bool 值（是否读成功）为 true。
    * 如果 chan 关闭前，通道内无元素(或元素已经被读完)，chan 内无值，接下来所有接收的值都会非阻塞直接成功，返回元素类型的零值，但是第二个 bool 值一直为 false。

## 示例
1. 写已经关闭的 chan
    ```go
    package main

    func main() {
        c := make(chan int, 3)
        close(c)
        c <- 1
    }
    // panic: send on closed channel
    // goroutine 1 [running]:
    // main.main()
    //         E:/code/dayDayUp/go/test/main.go:6 +0x45
    ```

2. 读已经关闭的 chan
以下示例为各数据类型的chan， 先存放一个元素， 然后关闭chan后， 连续读取三次
```go
package main

import "fmt"

func main() {
	chanInt := make(chan int, 5)
	chanInt <- 1
	close(chanInt)
	num, ok := <-chanInt
	fmt.Printf("num=%v, ok=%v\n", num, ok) // num=1, ok=true
	num, ok = <-chanInt
	fmt.Printf("num=%v, ok=%v\n", num, ok) // num=0, ok=false
	num, ok = <-chanInt
	fmt.Printf("num=%v, ok=%v\n", num, ok) // num=0, ok=false

	chanStr := make(chan string, 5)
	chanStr <- "str"
	close(chanStr)
	str, ok := <-chanStr
	fmt.Printf("str=%v, ok=%v\n", str, ok) // str=str, ok=true
	str, ok = <-chanStr
	fmt.Printf("str=%v, ok=%v\n", str, ok) // str=, ok=false
	str, ok = <-chanStr
	fmt.Printf("str=%v, ok=%v\n", str, ok) // str=, ok=false

	chanBool := make(chan bool, 5)
	chanBool <- true
	close(chanBool)
	bo, ok := <-chanBool
	fmt.Printf("bo=%v, ok=%v\n", bo, ok) // bo=true, ok=true
	bo, ok = <-chanBool
	fmt.Printf("bo=%v, ok=%v\n", bo, ok) // bo=false, ok=false
	bo, ok = <-chanBool
	fmt.Printf("bo=%v, ok=%v\n", bo, ok) // bo=false, ok=false

	type User struct {
		Name string
		age  int
	}
	chanUser := make(chan User, 5)
	chanUser <- User{Name: "dayDayUp", age: 1}
	close(chanUser)
	user, ok := <-chanUser
	fmt.Printf("user=%v, ok=%v\n", user, ok) // user={dayDayUp 1}, ok=true
	user, ok = <-chanUser
	fmt.Printf("user=%v, ok=%v\n", user, ok) // user={ 0}, ok=false
	user, ok = <-chanUser
	fmt.Printf("user=%v, ok=%v\n", user, ok) // user={ 0}, ok=false
}
```

## 解析
1. 为何写已关闭的chan, 会panic
查看`chan`源码
```go
// src/runtime/chan.go

/*
 * generic single channel send/recv
 * If block is not nil,
 * then the protocol will not
 * sleep but return if it could
 * not complete.
 *
 * sleep can wake up with g.param == nil
 * when a channel involved in the sleep has
 * been closed.  it is easiest to loop and re-run
 * the operation; we'll see that it's now closed.
 */
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
// ...
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}
// ...
}
```
`c.closed`表示通道是否关闭, 直接`panic`, panic的内容为`send on closed channel`

2. 为何读已关闭的chan, 能一直读到值
```go
// src/runtime/chan.go

// chanrecv receives on channel c and writes the received data to ep.
// ep may be nil, in which case received data is ignored.
// If block == false and no elements are available, returns (false, false).
// Otherwise, if c is closed, zeros *ep and returns (true, false).
// Otherwise, fills in *ep with an element and returns (true, true).
// A non-nil ep must point to the heap or the caller's stack.
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
    // ...
     
    // 管道已关闭
	if c.closed != 0 {
        // 管道中缓存为空
		if c.qcount == 0 {
			if raceenabled {
				raceacquire(c.raceaddr())
			}
			unlock(&c.lock)
            // ep为 val,ok := <- c 中的 val的地址
			if ep != nil {
                // 当ep不为空时, 获得该类型的零值
				typedmemclr(c.elemtype, ep)
			}
			return true, false
		}
		// The channel has been closed, but the channel's buffer have data.
	} else {
		// Just found waiting sender with not closed.
		if sg := c.sendq.dequeue(); sg != nil {
			// Found a waiting sender. If buffer is size 0, receive value
			// directly from sender. Otherwise, receive from head of queue
			// and add sender's value to the tail of the queue (both map to
			// the same buffer slot because the queue is full).
			recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
			return true, true
		}
	}
    // ...
}
```
其中, 核心内容为:
```go
if c.closed != 0 {
    // 管道中缓存为空
	if c.qcount == 0 {
		if raceenabled {
			raceacquire(c.raceaddr())
		}
		unlock(&c.lock)
        // ep为 val,ok := <- c 中的 val的地址
		if ep != nil {
            // 当ep不为空时, 获得该类型的零值
			typedmemclr(c.elemtype, ep)
		}
		return true, false
	}
}
```
即: 当通道已关闭, 且缓冲区为空时, 如果接收值的地址`ep`不为空, 则获得`该类型的零值`