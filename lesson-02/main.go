// === Lesson 02：变量与类型 — 零值、短声明、基本类型 ===
// 目标：掌握 Go 变量的声明方式，理解零值哲学，知道什么时候用 const。

package main

import "fmt"

func main() {
	// ── 声明方式 ──
	// 方式 1：var 声明，显式类型。适合需要在声明后赋值的场景（循环、条件分支内）。
	var name string = "Go"
	fmt.Println(name)

	// 方式 2：var 声明，类型推导。类型从右侧表达式自动推断。
	var version = "1.26.4"
	fmt.Println(version)

	// 方式 3：短声明 :=，只能在函数内使用。最常用。
	year := 2026
	fmt.Println(year)

	// ── 零值：未初始化的变量不是随机值，是"零值" —— Go 不会让你读到垃圾数据 ──
	var (
		i   int     // 0
		f   float64 // 0.0
		b   bool    // false
		s   string  // ""（空字符串，不是 nil）
		ptr *int    // nil
	)
	fmt.Printf("零值: int=%d float=%.1f bool=%v str=%q ptr=%v\n", i, f, b, s, ptr)

	// ── 基本类型 ──
	// int 的位数取决于平台（64位系统就是 int64），明确位数用 int32/int64。
	var small int32 = 32767 // 16位有符号范围，故意卡在 int16 上限
	var big int64 = 1 << 40 // 左移 40 位 = 1TB，int32 会溢出——但因为 big 是 int64，安全
	fmt.Printf("int32 max-ish: %d | int64 1<<40: %d\n", small, big)

	// float64 是 Go 浮点的默认类型，绝大多数场景用它没错。
	pi := 3.14159
	fmt.Printf("pi = %.5f\n", pi)

	// byte 是 uint8 的别名，rune 是 int32 的别名（代表 Unicode 码点）。
	var ch byte = 'A'        // ASCII 字母
	var cn rune = '中'        // 中文字符 = Unicode
	fmt.Printf("byte=%c(%d) | rune=%c(%d)\n", ch, ch, cn, cn)

	// ── const：编译期常量，不占运行时内存（每次用到都会内联为字面量） ──
	const MaxRetry = 3              // 无类型常量，赋值给 int/float 都行
	const Greeting = "hello " + "go" // 常量表达式在编译期求值
	fmt.Printf("const: MaxRetry=%d Greeting=%s\n", MaxRetry, Greeting)

	// ── 类型转换必须显式 ──
	// 下面这行编译不过：var x int = pi
	x := int(pi) // 显示截断小数部分 → 3
	fmt.Printf("int(pi) = %d\n", x)

	// ── 命名返回值：函数签名里声明返回变量，函数体内直接用 ──
	area, circum := circleInfo(5.0)
	fmt.Printf("圆的面积=%.2f, 周长=%.2f\n", area, circum)
}

// 多返回值——Go 没有异常机制，error 靠返回，大部分操作靠多返回值。
// 命名返回值 (area float64, circum float64) 在函数体可以直接用，最末 return 不必写变量名。
func circleInfo(radius float64) (area float64, circum float64) {
	area = 3.14159 * radius * radius
	circum = 2 * 3.14159 * radius
	return // naked return——命名返回值版本，等价于 return area, circum
}

// ── 练习 ──
// 1. 把 MaxRetry 改成 5，看编译速度没有任何变化（const 编译期替换，无运行时开销）。
// 2. 写一个函数 addAndMul(a, b int) (sum int, product int)，感受多返回值。
// 3. 声明 var uninit int，不赋值直接打印，确认它就是 0。
