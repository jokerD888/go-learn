// === Lesson 09：goroutine — 并发的最小单元 ===
// 目标：掌握 go 关键字、WaitGroup、理解 goroutine 调度模型。

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// ── goroutine：比线程更轻量的并发体 ──
	// go 关键字启动一个 goroutine，调度器将成千上万个 goroutine 映射到几个 OS 线程。
	// 启动一个 goroutine 只需要几 KB 栈空间，线程需要 ~1MB。

	// 错误示范：直接 go 启动，main 退出时 goroutine 全部被杀。
	// 下面这行大概率看不到输出——main 返回太快了。
	go fmt.Println("这条可能永远不会打印")

	// ── sync.WaitGroup：等一组 goroutine 全部完成 ──
	// 这是同步 goroutine 最基本的方式。
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1) // 计数器 +1，必须在 goroutine 外部调用
		// 闭包陷阱：i 是循环变量，每次迭代都被复用。
		// 用参数传递当前值，而非在闭包内引用 &i。
		go func(id int) {
			defer wg.Done() // 完成时计数器 -1，用 defer 保证一定执行
			time.Sleep(time.Duration(id*100) * time.Millisecond)
			fmt.Printf("goroutine %d 完成\n", id)
		}(i)
	}

	wg.Wait() // 阻塞直到计数器归零
	fmt.Println("全部 goroutine 完成\n")

	// ── 闭包陷阱演示 ──
	// 错误写法：所有 goroutine 共享同一个 i。
	fmt.Println("错误写法——闭包捕获循环变量:")
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Printf("  错误: i=%d (可能全是 3)\n", i)
		}()
	}
	time.Sleep(100 * time.Millisecond)

	// Go 1.22+ 已修复此问题——循环变量每轮迭代独立。
	// 但显式传参仍是更清晰的写法，也兼容旧版本。
	fmt.Println("Go 1.22+ 循环变量已自动独立，下面是对的:")
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Printf("  正确: i=%d\n", i)
		}()
	}
	time.Sleep(100 * time.Millisecond)

	// ── goroutine 数量不是越多越好 ──
	// goroutine 很轻但调度有开销。通常用 worker pool 限制并发数。
	// 标准方案：带缓冲 channel 做信号量（下节课讲）。

	// ── 数据竞争：两个 goroutine 同时写一个变量 ──
	// go run -race ./lesson-09/ 可以检测出竞争。
	// Go 的 race detector 是运行时检测，对性能有约 10x 影响，开发/测试时用。
	counter := demoRaceCondition()
	fmt.Printf("并发累加 1000 次，结果: %d (大概率 < 1000, 因为覆盖写入)\n", counter)
}

// 有数据竞争的累加器——并发写同一个 int 导致丢失更新。
func demoRaceCondition() int {
	var wg sync.WaitGroup
	count := 0
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count++ // 竞争：读-改-写三步不是原子的
		}()
	}
	wg.Wait()
	return count
}

// ── goroutine 三定律 ──
// 1. goroutine 的退出只能由自己 return（或 panic），不能从外部杀死它。
//    想"杀" goroutine？用 channel 通知它自己退出（下节课）。
// 2. main 函数返回 → 所有 goroutine 一起终止，没有"等待"。
//    所以主 goroutine 必须用 WaitGroup/channel 等子 goroutine。
// 3. goroutine 不能有返回值，通过 channel 传回结果。

// ── 练习 ──
// 1. 用 go run -race ./lesson-09/ 运行本课，观察 race detector 输出。
// 2. 把 demoRaceCondition 的 goroutine 数量改成 10000，观察结果偏离程度。
// 3. 用 sync.WaitGroup 并发下载 3 个"网页"（用 time.Sleep 模拟），等全部完成再输出。
