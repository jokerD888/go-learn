// === Lesson 08：错误处理 — error 接口、包装、errors.Is/As ===
// 目标：掌握 Go 的错误处理模式，理解 sentinel error 和 error wrapping。

package main

import (
	"errors"
	"fmt"
	"os"
)

// ── error 是内置接口，只有一个方法 Error() string ──
// type error interface { Error() string }
// 任何实现了 Error() 的类型都是 error。

// ── 方式 1：fmt.Errorf 直接创建 ──
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零: a=%.1f", a)
	}
	return a / b, nil
}

// ── 方式 2：errors.New —— 只有一条消息，不需要格式化时用 ──
var ErrNotFound = errors.New("未找到") // sentinel error：包级别的 error 变量

// ── 方式 3：自定义 error 类型 —— 携带结构化信息 ──
// 惯例：error 类型名以 Error 结尾。
type ValidationError struct {
	Field string
	Value string
	Msg   string
}

// Error() 方法实现 error 接口。
// 指针接收者：防止每次返回都复制，也让 nil 检查更直观。
func (e *ValidationError) Error() string {
	return fmt.Sprintf("校验失败: %s=%q, %s", e.Field, e.Value, e.Msg)
}

func validateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Value: email, Msg: "不能为空"}
	}
	return nil
}

func main() {
	// ── 基础用法：调用 → 检查 err != nil → 处理 ──
	if _, err := divide(10, 0); err != nil {
		fmt.Println("divide:", err)
	}

	// ── 自定义 error 的断言 ──
	err := validateEmail("")
	if err != nil {
		fmt.Println("validateEmail:", err)
		// 断言具体类型，获取结构化信息。
		// 注意：用 *ValidationError 不是 ValidationError，因为 Error() 绑定在指针上。
		var valErr *ValidationError
		if errors.As(err, &valErr) { // errors.As 遍历错误链，找到匹配类型
			fmt.Printf("  字段=%s, 值=%q\n", valErr.Field, valErr.Value)
		}
	}

	// ── error wrapping：用 %w 包装错误，保留原始错误 ──
	// Go 1.13+ 的 %w 将原始 error 嵌入新 error，形成"错误链"。
	// 上层可以解包拿到原始错误，同时添加上下文。
	_, fileErr := readConfig("config.yaml")
	if fileErr != nil {
		fmt.Println("readConfig:", fileErr)

		// errors.Is 检查错误链上是否有匹配的 sentinel error。
		// 这是比 err == ErrNotFound 更安全的写法——它能穿透 %w 包装。
		if errors.Is(fileErr, os.ErrNotExist) {
			fmt.Println("  文件不存在，请检查路径")
		}
	}

	// ── errors.Is vs errors.As ──
	// Is：检查错误链上是否有"某个特定的 error 值"（sentinel error）。
	// As：检查错误链上是否有"某个类型的 error"（custom error type）。

	// sentinel error 例子
	_ = runApp()
}

func readConfig(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// %w 包装：把底层错误（os.ErrNotExist）和业务上下文拼在一起。
		return nil, fmt.Errorf("读取配置失败 %s: %w", path, err)
	}
	return data, nil
}

// ── 实战：多层错误包装 ──

// 业务 sentinel error。
var (
	ErrDBConnection = errors.New("数据库连接失败")
	ErrTimeout      = errors.New("操作超时")
)

func queryUser(id int) error {
	// 模拟底层返回 sentinel error。
	baseErr := ErrDBConnection

	// 包装两层，每层加更多上下文。
	inner := fmt.Errorf("查询用户(id=%d)失败: %w", id, baseErr)
	return fmt.Errorf("user.GetById: %w", inner)
}

func runApp() error {
	err := queryUser(42)
	if err == nil {
		return nil
	}

	// errors.Is 遍历整个错误链，找到最底层匹配的 sentinel error。
	if errors.Is(err, ErrDBConnection) {
		fmt.Println("检测到数据库故障，请检查连接")
	}
	// errors.As 提取错误链上的具体类型。
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Println("校验错误:", valErr.Field)
	}

	// %v 打印完整链（Go 1.13+），每层一行——非常适合调试。
	fmt.Printf("\n完整错误链:\n%v\n", err)
	return err
}

// ── 错误处理准则 ──
// 1. 调用方必须处理 err，不允许忽略。忽略唯一的例外：println 这种永远不失败的。
// 2. 不要预先声明 err 变量在函数顶部——在需要检查的地方用 if err := ... 限定作用域。
// 3. 包装错误时描述"做什么失败了"，底层错误用 %w 传入。不要重复底层错误的信息。
// 4. 只在包边界需要包装（repository → service → handler），内部调用链不重复包装。
// 5. sentinel error 命名以 Err 开头，变量名要具体（ErrNotFound 好过 ErrGeneric）。

// ── 练习 ──
// 1. 定义一个自定义 error 类型 QueryError{Query string, Err error}，实现 Error() 和 Unwrap()。
// 2. 用 fmt.Errorf 包装三次，然后用 errors.Is 和 errors.As 分别追溯。
// 3. 把 divide 改成返回自定义 DivisionError，携带被除数和除数。
