// === Lesson 04：函数 — 多返回值、命名返回、闭包 ===
// 目标：掌握 Go 函数的设计模式，理解闭包捕获的本质。

package main

import (
	"errors"
	"fmt"
	"strings"
)

func main() {
	// ── 多返回值 —— Go 的"异常"机制 ──
	// Go 没有 try/catch。函数通过返回 error 表示失败，调用方必须检查。
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("10 / 2 = %.1f\n", result)
	}

	// 除零返回 error，而不是 panic——可控。
	if _, err := divide(10, 0); err != nil {
		fmt.Println("除零结果:", err)
	}

	// ── 变长参数 ... ──
	// ...int 在函数内部是一个 []int 切片，调用时可以传任意个 int。
	fmt.Println("sum(1,2,3,4,5) =", sumAll(1, 2, 3, 4, 5))
	// 已有切片传给变长参数，加 ... 展开。
	nums := []int{10, 20, 30}
	fmt.Println("sumAll(nums...) =", sumAll(nums...))

	// ── 函数是一等公民：可以赋值给变量、当参数传、当返回值 ──
	op := add // 函数赋值给变量，不加括号
	fmt.Println("op(3, 4) =", op(3, 4))

	// 高阶函数：函数接受函数作为参数
	r := compute(10, 2, add)
	fmt.Println("compute(10, 2, add) =", r)

	// ── 闭包（closure）：函数捕获外部变量 ──
	// 闭包 = 函数 + 它引用的自由变量。Go 的闭包捕获的是变量的引用，不是值的拷贝。
	counter := makeCounter()  // counter 是一个 func() int
	fmt.Println(counter())    // 1
	fmt.Println(counter())    // 2
	fmt.Println(counter())    // 3
	// 每次调用 counter()，它操作的是 makeCounter 内的同一个 count 变量。

	// ── 闭包陷阱：循环中启用 goroutine 时，捕获循环变量要用参数传递 ──
	// 这是一个经典错误模式的正确解法。
	funcs := make([]func() string, 3)
	for i := 0; i < 3; i++ {
		// 正确：用参数传递当前值，而非直接捕获 &i。
		funcs[i] = func(val int) func() string {
			return func() string {
				return fmt.Sprintf("val=%d", val)
			}
		}(i)
	}
	for _, fn := range funcs {
		fmt.Println(fn())
	}

	// ── errors 包：标准库的错误处理 ──
	validate("")       // 触发错误
	validate("hello")  // 正常
}

// divide 返回两个值：(结果, 错误)。这是 Go 的习惯，不是额外的"功能"。
func divide(a, b float64) (float64, error) {
	if b == 0 {
		// errors.New 创建最简单的 error，不含调用栈。
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil // nil 表示没有错误
}

// ...int = 变长参数，函数内部 nums 是 []int。
func sumAll(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 普通函数。
func add(a, b int) int { return a + b }

// 高阶函数：第三个参数是 func(int, int) int 类型。
func compute(a, b int, fn func(int, int) int) int {
	return fn(a, b)
}

// makeCounter 返回一个闭包——返回的函数"记住"了 count。
// count 存储在堆上，即使 makeCounter 已返回仍然存活。
func makeCounter() func() int {
	count := 0
	return func() int {
		count++ // 捕获的是 &count，不是 count 的副本
		return count
	}
}

// strings.TrimSpace 是标准库提供的，不需要自己写。
func validate(name string) {
	if strings.TrimSpace(name) == "" {
		fmt.Println("validate: 名字不能为空")
		return
	}
	fmt.Printf("validate: %q 通过\n", name)
}

// ── 练习 ──
// 1. 自己写一个闭包 makeAdder(base int) func(int) int，每次调用加 base。
// 2. 对比"命名返回值"和"非命名返回值"两种写法。
// 3. 用 errors.New 定义两个错误，再用 fmt.Errorf("包装: %w", err) 包装一层。
