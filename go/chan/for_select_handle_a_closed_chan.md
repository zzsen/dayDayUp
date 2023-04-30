# for 循环 select 里, 如果通道已关闭, 会怎样? 如果 select 里只有一个 case 呢?

## 答案

- for 循环 select 里, 如果其中一个通道已经关闭, 则每次都会执行这个 case
- 如果只有一个 case, 且 case 里的通道已关闭, 则会出现死循环

## 示例

### 1. for 循环 select 里, 其中一个通道已关闭

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- false
		close(ch)
	}()

	for {
		select {
		case val, ok := <-ch:
			fmt.Printf("case chan, val=%v, ok=%v\n", val, ok)
			time.Sleep(1 * time.Second)
		default:
			fmt.Println("case default")
			time.Sleep(1 * time.Second)
		}
	}
}
// case default
// case default
// case chan, val=true, ok=true
// case chan, val=false, ok=false
// case chan, val=false, ok=false
// case chan, val=false, ok=false
// case chan, val=false, ok=false
// case chan, val=false, ok=false
```

由于读已关闭的通道, 可以一直读到值, 故会执行通道的 case

### 2. for 循环 select 里, 其中一个通道已关闭, 如何才能不读关闭的通道

将通道置空即可

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- true
		close(ch)
	}()

	for {
		select {
		case val, ok := <-ch:
			fmt.Printf("case chan, val=%v, ok=%v\n", val, ok)
			time.Sleep(1 * time.Second)
			if !ok {
				ch = nil
			}
		default:
			fmt.Println("case default")
			time.Sleep(1 * time.Second)
		}
	}
}
// case default
// case default
// case chan, val=true, ok=true
// case chan, val=false, ok=false
// case default
// case default
// case default
// case default
```

将通道置空后, 读未初始化的通道, 会发生阻塞, 由于 select 会跳过阻塞的 case, 故不会再去读已关闭(关闭后置空)的通道

### 3. 如果只有一个 case 呢

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- true
		close(ch)
	}()

	for {
		select {
		case val, ok := <-ch:
			fmt.Printf("case chan, val=%v, ok=%v\n", val, ok)
			time.Sleep(1 * time.Second)
		}
	}
}
// case chan, val=true, ok=true
// case chan, val=false, ok=false
// case chan, val=false, ok=false
// case chan, val=false, ok=false
// case chan, val=false, ok=false
```

由于读已关闭的通道, 可以一直读到值, 故会执行通道的 case

### 4. 只有一个 case, 且已关闭, 且置为 nil

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- true
		close(ch)
	}()

	for {
		select {
		case val, ok := <-ch:
			fmt.Printf("case chan, val=%v, ok=%v\n", val, ok)
			time.Sleep(1 * time.Second)
			if !ok {
				ch = nil
			}
		}
	}
}
// case chan, val=true, ok=true
// case chan, val=false, ok=false
// fatal error: all goroutines are asleep - deadlock!

// goroutine 1 [chan receive (nil chan)]:
// main.main()
//         D:/projects/dayDayUp/go/main.go:19 +0xa8
```

- 第一次读取到的是通道关闭后剩余的元素
- 第二次读已关闭的通道, 能读到零值, 但 ok 值为 false
- 第三次通道已经置为 nil, main 协程阻塞, 整个进程无其他协程, 故死锁

故 select 最好有 default, 以预防死锁发生.
