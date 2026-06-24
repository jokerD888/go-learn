// === Lesson 09：goroutine — 并发的最小单元 ===
// 目标：掌握 go 关键字、WaitGroup、闭包陷阱、调度模型、race condition。

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 1. goroutine 基础 + WaitGroup ===")
	// goroutine：Go 的轻量级用户态线程。启动只需 ~4KB 栈，OS 线程 ~1MB。
	// go 关键字启动后，调度器把成千上万个 goroutine 映射到 GOMAXPROCS 个 OS 线程。

	// 错误示范：直接 go 启动无同步，main 退出时可能看不到输出。
	go fmt.Println("这条可能永远不会打印——main 返回太快")

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1) // 必须在 goroutine 外部 Add，否则计数器可能尚未增加就 Wait 返回
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id*100) * time.Millisecond)
			fmt.Printf("goroutine %d 完成\n", id)
		}(i) // 参数传值，避免闭包捕获循环变量
	}
	wg.Wait() // 阻塞直到计数器归零
	fmt.Println("全部 goroutine 完成\n")

	// ── 2. 闭包陷阱：可验证的对比实验 ──
	fmt.Println("=== 2. 闭包陷阱 ===")
	// 错误写法：闭包捕获的是变量 i 的引用，循环结束后 i==3，所有 goroutine 读到同一个 3。
	fmt.Println("错误写法（捕获循环变量 i）：")
	var wg2 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			fmt.Printf("  错误: i=%d\n", i) // 大概率全是 3
		}()
	}
	wg2.Wait() // WaitGroup 等所有 goroutine 完成再继续——不用 sleep 瞎等

	// 正确写法：参数传值，每个 goroutine 拿到的是当时的副本。
	fmt.Println("正确写法（参数传递）：")
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			fmt.Printf("  正确: id=%d\n", id)
		}(i)
	}
	wg2.Wait()

	// Go 1.22+ 循环变量每次迭代独立，闭包直接捕获也安全。但显式传参兼容旧版，语义更清晰。

	// ── 3. GOMAXPROCS：并发 ≠ 并行 ──
	fmt.Println("\n=== 3. GOMAXPROCS ===")
	runtime.GOMAXPROCS(1) // 只用 1 个 OS 线程——所有 goroutine 都是"并发"而非"并行"
	t0 := time.Now()
	var wg3 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg3.Add(1)
		go func(id int) {
			defer wg3.Done()
			time.Sleep(50 * time.Millisecond) // 模拟计算
		}(i)
	}
	wg3.Wait()
	// GOMAXPROCS=1 意味着调度的延迟，执行时间接近 3×50ms（但实际可能有微小出入）
	fmt.Printf("GOMAXPROCS=1 耗时: %v\n", time.Since(t0))

	// 如果 >=3 个 OS 线程并行：时间接近 50ms
	runtime.GOMAXPROCS(runtime.NumCPU()) // 恢复默认
	// 思考：为什么 GOMAXPROCS=1 时，3 个 goroutine 不能同时执行？
	// 因为只有 1 个 OS 线程，goroutine 必须轮流获得调度。多线程才可能真正"同时"跑。

	// ── 4. race condition + 解法 ──
	fmt.Println("\n=== 4. race condition ===")
	// 无保护并发写入——运行 go run -race main.go 可检测。
	unsafe := demoRaceUnsafe()
	fmt.Printf("无保护并发加 1000 次，结果: %d（大概率 < 1000，覆盖写入导致丢失）\n", unsafe)

	// 解法 1：Mutex
	safeMutex := demoRaceMutex()
	fmt.Printf("Mutex 保护后: %d\n", safeMutex)

	// 解法 2：atomic
	safeAtomic := demoRaceAtomic()
	fmt.Printf("atomic 保护后: %d\n", safeAtomic)

	// ── 5. goroutine 泄漏 —— 最常踩的坑 ──
	fmt.Println("\n=== 5. goroutine 泄漏 ===")
	leakExample() // 注意注释：这段代码演示了泄漏，不会 block 但会留下孤儿 goroutine
}

// ── race condition 示例 ──

func demoRaceUnsafe() int {
	var wg sync.WaitGroup
	count := 0
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count++ // 读-改-写 非原子 → 并发覆盖
		}()
	}
	wg.Wait()
	return count
}

// 解法 1：sync.Mutex 保护临界区
func demoRaceMutex() int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			count++ // 临界区受保护
			mu.Unlock()
		}()
	}
	wg.Wait()
	return count
}

// 解法 2：atomic 无锁加法
func demoRaceAtomic() int32 {
	var wg sync.WaitGroup
	var count int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt32(&count, 1) // 硬件级别原子操作
		}()
	}
	wg.Wait()
	return count
}

// ── goroutine 泄漏 ──
// 向无缓冲 channel 发送但没有接收者 → goroutine 永远阻塞 → 无法被 GC → 内存泄漏。
func leakExample() {
	timeout := time.After(50 * time.Millisecond)
	done := make(chan bool)
	go func() {
		// 模拟耗时 200ms，但外部 50ms 超时后不再读 done → 这个 goroutine 永远不会 return。
		time.Sleep(200 * time.Millisecond)
		done <- true // 无缓冲，无接收者，永久阻塞
		fmt.Println("〈这行永远不会输出〉")
	}()
	select {
	case <-done:
		fmt.Println("完成")
	case <-timeout:
		fmt.Println("超时！但那个 goroutine 泄漏了——它还在等 done 的接收者")
	}
	time.Sleep(50 * time.Millisecond) // 让 select 走完
	// 生产环境正确做法：不等待的 channel 设成带缓冲（done := make(chan bool, 1)）
	// 或者在对方不再读取时用 select+default 非阻塞发送。
}

// ── goroutine 四定律 ──
// 1. goroutine 只能自己 return（或 panic）退出，不能从外部"杀死"。
//    → 想控制退出？用 context 或 channel 通知它自己 return。
// 2. main 返回 → 所有 goroutine 直接终止，没有"优雅退出"。
//    → 主 goroutine 必须同步：WaitGroup / channel / context。
// 3. goroutine 无返回值 → 通过 channel 传回结果。
// 4. goroutine 会泄漏。无缓冲 channel 上永久阻塞的 goroutine 永远不会被 GC。
//    → 关注 goroutine 的生命周期，确保每个都有退出路径。

// ── 练习 ──
// 1. go run -race main.go 运行，观察 race detector 对 demoRaceUnsafe 的输出。
// 2. 把 demoRaceUnsafe 的并发数改成 10000，对比 Mutex/atomic 的结果。
// 3. 三个"下载任务"（time.Sleep 模拟），用 WaitGroup 收集全部完成再输出。
// 4. 验证 goroutine 泄漏：试着用 taskkill 查进程，或用 benchmark 看内存增长（进阶）。
