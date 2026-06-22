// lesson-11 的测试文件。文件名以 _test.go 结尾，go test 才会识别。

package main

import (
	"fmt"
	"testing"
)

// ── Table-driven test：Go 测试的标准模式 ──
// 把输入和期望输出放进 struct 切片，遍历验证。
// 好处：新增用例只需加一行 struct{}，不用复制测试逻辑。

func TestAdd(t *testing.T) {
	tests := []struct { // 匿名 struct 定义测试表格
		name     string // 用例名称，方便定位失败
		a, b     int
		expected int
	}{
		{"正数相加", 1, 2, 3},
		{"零值", 0, 5, 5},
		{"负数", -1, -2, -3},
		{"正负混合", 10, -3, 7},
	}
	for _, tt := range tests {
		// t.Run 将每个用例作为独立子测试运行。
		// go test -run "TestAdd/正数相加" 可以只跑一个子用例。
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.a, tt.b); got != tt.expected {
				t.Errorf("Add(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

// 测试包含 error 返回的函数。
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool // 是否期望返回 error
	}{
		{"正常除法", 10, 2, 5, false},
		{"除零", 10, 0, 0, true},
		{"负数", 10, -2, -5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Divide(%f,%f) error=%v, wantErr=%v", tt.a, tt.b, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Divide(%f,%f) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) {
		t.Error("应该包含 2")
	}
	if Contains([]string{"a", "b"}, "c") {
		t.Error("不应该包含 c")
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"普通字符串", "hello", "olleh"},
		{"含中文", "Go 语言", "言语 oG"},
		{"空字符串", "", ""},
		{"单字符", "a", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.input); got != tt.want {
				t.Errorf("Reverse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── Benchmark：性能测试 ──
// 函数以 Benchmark 开头，参数是 *testing.B。
// b.N 由测试框架自动调整，直到测试稳定运行 1 秒。

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(1, 2)
	}
}

func BenchmarkContains10(b *testing.B) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer() // 排除准备数据的耗时
	for i := 0; i < b.N; i++ {
		Contains(data, 7)
	}
}

// 子 benchmark：比较不同规模数据。
func BenchmarkContainsScale(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Contains(data, size-1) // 找最后一个元素——最坏情况
			}
		})
	}
}

// ── 测试规范 ──
// 1. 测试文件与源文件同目录，名为 *_test.go。
// 2. 函数签名：TestXxx(t *testing.T)，BenchmarkXxx(b *testing.B)。
// 3. 每个测试独立——不依赖全局状态、不依赖执行顺序。
// 4. table-driven 是默认风格，不要一个函数测一个用例。
// 5. error message 要包含 got/want，用 %q 打印字符串（带引号，空格可见）。
// 6. 不要忽略 coverage 报告。至少核心逻辑 >80%。
