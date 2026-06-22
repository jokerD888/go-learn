// === Lesson 10：channel — goroutine 间的通信 ===
// 目标：掌握 channel 的创建、发送、接收、select、close 模式。
// Go 哲学: "不要通过共享内存来通信，通过通信来共享内存。"

package main

import (
	"fmt"
	"time"
)

func main() {
	// ── channel 创建与基本操作 ──
	// make(chan T) —— 无缓冲 channel，发送方阻塞直到接收方就绪。
	// 无缓冲 = 同步 channel，同时完成发送和接收。

	ch := make(chan string) // 无缓冲

	// 发送方在单独的 goroutine 中，因为无缓冲发送会阻塞直到有人收。
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- "hello from goroutine" // 发送：接收方未就绪时阻塞
	}()

	// 接收方阻塞等数据到来。
	msg := <-ch
	fmt.Println("收到:", msg)

	// ── 带缓冲 channel：容量满之前不阻塞 ──
	// make(chan T, N) —— N 是缓冲区大小。
	buf := make(chan int, 3) // 缓冲 3 个
	buf <- 1                 // 不阻塞
	buf <- 2                 // 不阻塞
	buf <- 3                 // 不阻塞
	// buf <- 4               // 阻塞——缓冲区满，没有接收方
	fmt.Println("缓冲 channel 已有 3 个元素")
	close(buf) // 关闭后仍可读取剩余数据
	// range channel 读取到 close 为止。
	for v := range buf {
		fmt.Println("  range 读取:", v)
	}

	// ── select：同时监听多个 channel ──
	// select 随机选择一个就绪的 case 执行，没有就绪则阻塞。
	// 这是 Go 并发编程的核心控制结构。
	fmt.Println("\n=== select 演示 ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "通道1 先到"
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "通道2 后到"
	}()

	// select 会先收到 ch1（50ms < 100ms）。
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Println("收到:", msg1)
		case msg2 := <-ch2:
			fmt.Println("收到:", msg2)
		case <-time.After(200 * time.Millisecond): // 超时控制
			fmt.Println("超时！")
		}
	}

	// ── 生产者-消费者模式 ──
	fmt.Println("\n=== 生产者-消费者 ===")
	jobs := make(chan int, 5)
	results := make(chan int, 5)

	// 启动 3 个 worker。
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// 生产 5 个任务。
	for j := 1; j <= 5; j++ {
		jobs <- j
	}
	close(jobs) // 生产完关闭，worker 的 range 自动结束

	// 收集结果。
	for r := 1; r <= 5; r++ {
		res := <-results
		fmt.Println("结果:", res)
	}

	// ── select default：非阻塞操作 ──
	fmt.Println("\n=== 非阻塞发送 ===")
	nonBlockCh := make(chan int, 1)
	nonBlockCh <- 1 // 先填满
	select {
	case nonBlockCh <- 2:
		fmt.Println("发送成功")
	default:
		fmt.Println("channel 满了，跳过发送") // 不会阻塞
	}

	// ── 用 channel 做信号通知 ──
	fmt.Println("\n=== done channel 模式 ===")
	done := make(chan struct{}) // struct{} 零字节，纯信号
	go func() {
		fmt.Println("  工作中...")
		time.Sleep(200 * time.Millisecond)
		close(done) // close channel = 广播"完成了"，所有接收者立即返回
	}()
	<-done // 阻塞等到关闭
	fmt.Println("  收到完成信号")

	// ── time.Tick vs time.After ──
	// Tick 返回一个循环发送的 channel——但它不会自动回收，慎用。
	// After 返回一个发送一次的 channel，常用于超时。
}

// worker 从 jobs 读取，计算结果写入 results。
func worker(id int, jobs <-chan int, results chan<- int) {
	// <-chan 表示只读 channel，chan<- 表示只写 channel。
	// 编译期保证不会误操作方向——这是 Go 的类型安全。
	for j := range jobs { // range channel 在 close 后自动退出
		fmt.Printf("  worker %d 处理任务 %d\n", id, j)
		time.Sleep(100 * time.Millisecond) // 模拟耗时
		results <- j * 2
	}
}

// ── channel 设计原则 ──
// 1. 谁创建，谁关闭。接收方不应该关闭 channel。
// 2. 向已关闭的 channel 发送 → panic。这是最常见的并发 bug 之一。
// 3. 从已关闭的 channel 接收 → 返回零值 + false（用 v, ok := <-ch 检测）。
// 4. 永远不要关闭一个可能有多个发送方的 channel——用 sync.WaitGroup 替代。
// 5. nil channel 的发送和接收永久阻塞，select 中 nil case 被忽略——可利用。

// ── 练习 ──
// 1. 写一个 fan-in：两个 channel 合并到一个 channel，用 select 实现。
// 2. 用带缓冲 channel 实现一个固定大小的 worker pool（信号量模式）。
// 3. 用 done channel 实现 goroutine 的优雅退出（不关闭 jobs，收 done 后不再发送新任务）。
