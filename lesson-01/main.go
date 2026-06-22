// === Lesson 01：Hello Go — 工具链、包、编译运行 ===
// 目标：写一个 Go 程序，理解 package/module/import 的关系，能编译+运行。

package main

// import 导入了包就叫 "依赖"，不导入不让用——编译时就会报错。
import (
	"fmt"  // 格式化输出，标准库之一
	"math" // 也是标准库
)

// main() 是程序的唯一入口，每个可执行程序必须有且仅有一个。
// Go 没有 OOP 的 class，所有代码都写在"包"里，main 包 = 可执行程序。
func main() {
	// := 短声明：声明变量 + 赋值，类型自动推导。Go 里面最常用的声明方式。
	greeting := "Hello, Go 1.26"
	fmt.Println(greeting)

	// math.Pow 返回 float64，Go 的数值类型严格，不会隐式转换。
	// 这就是 Go 的"显式"哲学——你知道每一行的确切行为。
	power := math.Pow(2, 10) // 2 的 10 次方
	fmt.Printf("2^10 = %.0f\n", power)

	// Sprintf 格式化到字符串，不输出到控制台。
	report := fmt.Sprintf("%s | computed: 2^10=%.0f", greeting, power)
	fmt.Println(report)
}

// ── 练习 ──
// 1. 把 greeting 改成你的名字，重新 go run .
// 2. 加一行 fmt.Println(math.Sqrt(100))，感受一下 import 的作用。
// 3. 试试把 import "math" 删了，看编译器报什么错。
