// === Lesson 03：控制流 — if/for/switch、defer 初识 ===
// 目标：理解 Go 控制流的简洁设计，以及 defer 的用途。

package main

import (
	"fmt"
	"os"
)

func main() {
	// ── if：条件不加括号，花括号必须写 ──
	// 允许在 if 前加一个简短语句（如 err 赋值），作用域限定在 if/else 内。
	if f, err := os.Open("not-exist.txt"); err != nil {
		// err != nil 是 Go 的错误处理惯例，绝不抛异常，返回 error 给你判断。
		fmt.Println("预期中的错误:", err)
	} else {
		// f 只在 if/else 块内可见——变量作用域严格控制，防止误用。
		defer f.Close()
	}

	// ── for 是 Go 唯一的循环关键字，没有 while ──
	// 形式 1：C 风格三段式
	sum := 0
	for i := 1; i <= 10; i++ {
		sum += i
	}
	fmt.Println("1+2+...+10 =", sum)

	// 形式 2：只有条件 = while
	// Go 没有 while，直接用 for + 条件即可。
	n := 1024
	for n > 1 {
		n /= 2
	}
	fmt.Printf("1024 不断 /2 直到 <=1: %d\n", n)

	// 形式 3：死循环 = for {}
	// 单写 for {} 就是死循环，break 退出。常用于服务器主循环。

	// 形式 4：range 遍历——最常用
	fruits := []string{"apple", "banana", "cherry"}
	for idx, fruit := range fruits {
		fmt.Printf("fruits[%d] = %s\n", idx, fruit)
	}
	// Go 不允许声明了变量不用。如果不要 idx，用 _ 占位：for _, fruit := range fruits

	// ── switch：不需要 break，默认只执行一个分支 ──
	// 不会 fall-through（和 C 不同），反而省了 break 的麻烦。
	os := "linux"
	switch os {
	case "windows":
		fmt.Println("你在 Windows 上——路径用 \\")
	case "linux", "darwin": // 多个条件用逗号
		fmt.Println("你在 Unix 系统上——路径用 /")
	default:
		fmt.Println("未知操作系统")
	}

	// switch 也可以不跟表达式，每个 case 是条件表达式。
	score := 85
	switch {
	case score >= 90:
		fmt.Println("A")
	case score >= 80:
		fmt.Println("B")
	case score >= 70:
		fmt.Println("C")
	default:
		fmt.Println("继续努力")
	}

	// ── defer：函数返回前一定执行（LIFO 顺序）——相当于 C 的 "清理栈" ──
	// 资源释放的标配：打开后立刻 defer Close()，不用记住在哪关。
	exampleDefer()
}

// defer 的多层执行顺序：后进先出（LIFO），像堆盘子。
// defer 的参数在声明时就求值了，不是最后执行时才求值。
func exampleDefer() {
	fmt.Println("--- defer 演示 ---")
	defer fmt.Println("第1个 defer（最先声明）")
	defer fmt.Println("第2个 defer（第二个声明）")
	defer fmt.Println("第3个 defer（最后声明）")
	fmt.Println("函数体执行中...")
	// 输出顺序：函数体 → 第3个 → 第2个 → 第1个（LIFO）
}

// ── 练习 ──
// 1. 用 for 打印 1-100 之间 3 的倍数。
// 2. 写一个 switch 判断 weekday，周一输出 "开工"，周六日输出 "躺平"- 用 fallthrough 试试。
// 3. 多个 defer 同时存在，猜猜顺序，跑一次验证。
