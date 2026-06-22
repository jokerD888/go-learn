// === Lesson 11：测试 — table-driven、benchmark、go test ===
// 目标：掌握 Go 内置测试框架，写出 table-driven 测试和 benchmark。

// 运行方式：
//   go run ./lesson-11/                — 跑演示
//   go test -v ./lesson-11/            — 跑测试
//   go test -run TestAdd ./lesson-11/  — 跑指定测试
//   go test -bench . ./lesson-11/      — 跑性能测试
//   go test -cover ./lesson-11/        — 查看覆盖率
//   go test -race ./lesson-11/         — 竞态检测

package main

import "fmt"

// ── 被测试的业务代码 ──

func Add(a, b int) int {
	return a + b
}

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零")
	}
	return a / b, nil
}

// Contains 检查切片是否包含目标值。
// 泛型版本，T 必须是可比较的类型。
func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// Reverse 反转字符串。
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	fmt.Println("Add(1,2) =", Add(1, 2))
	if r, err := Divide(10, 2); err == nil {
		fmt.Println("Divide(10,2) =", r)
	}
	fmt.Println(`Contains([1,2,3], 2) =`, Contains([]int{1, 2, 3}, 2))
	fmt.Println(`Reverse("Go 语言") =`, Reverse("Go 语言"))
	fmt.Println("\n运行 go test -v ./lesson-11/ 跑测试")
	fmt.Println("运行 go test -bench . ./lesson-11/ 跑基准测试")
}
