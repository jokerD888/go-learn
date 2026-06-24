// === Lesson 08：错误处理 — error 接口、构造方式、包装与解包、errors.Is/As ===
// 目标：掌握 Go 中创建、传递、包装和判断错误的完整体系。

package main

import (
	"errors"
	"fmt"
	"os"
)

// ── 1. error 接口：Go 最简单的内置接口 ──
// type error interface { Error() string }
// nil 表示无错误。非 nil → 有错误，调用方必须处理。

// ── 2. 创建 error 的三种方式 ──

// 方式 ①：errors.New — 纯字符串，不格式化，返回指针值可做 sentinel
var ErrNotFound = errors.New("未找到")

// 方式 ②：fmt.Errorf — 可带格式化参数
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零: a=%.1f, b=%.1f", a, b)
	}
	return a / b, nil
}

// 方式 ③：自定义 error 类型 — 携带结构化字段
// 自定义类型名以 Error 结尾（Go 惯例）。
type ValidationError struct {
	Field string
	Value string
	Msg   string
}

// 接收者用指针 *ValidationError，避免每次返回 error 时发生值拷贝。
// errors.As 在错误链上找到 *ValidationError 后，直接把值赋给传入的二级指针。
func (e *ValidationError) Error() string {
	return fmt.Sprintf("校验失败: %s=%q, %s", e.Field, e.Value, e.Msg)
}

func validateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Value: email, Msg: "不能为空"}
	}
	return nil
}

// ── 3. Unwrap：让自定义类型加入错误链 ──
// 实现 Unwrap() error 后，errors.Is / errors.As 就能穿透该类型继续往下找。
// fmt.Errorf 的 %w 自动生成的包装器内部也实现了 Unwrap。
type QueryError struct {
	Query  string
	Err    error // 包裹的底层错误
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("查询 %q 失败: %v", e.Query, e.Err)
}

// Unwrap 返回内层错误，errors.Is/As 靠它遍历链。
func (e *QueryError) Unwrap() error {
	return e.Err
}

func main() {
	fmt.Println("=== 1. 三种创建方式 ===")
	// ① errors.New
	fmt.Println("ErrNotFound:", ErrNotFound)

	// ② fmt.Errorf
	if _, err := divide(10, 0); err != nil {
		fmt.Println("divide:", err)
	}

	// ③ 自定义类型 + errors.As 提取
	err := validateEmail("")
	if err != nil {
		fmt.Println("validateEmail:", err)
		var valErr *ValidationError           // 用指针 *ValidationError，因为 Error() 在指针上
		if errors.As(err, &valErr) {          // As 遍历错误链，找到就写入 valErr
			fmt.Printf("  → 字段=%s, 值=%q\n", valErr.Field, valErr.Value)
		}
	}

	// ── 4. %w 包装与 errors.Is — 穿透多层找回 sentinel error ──
	fmt.Println("\n=== 2. 错误包装 + errors.Is ===")
	_, fileErr := readConfig("config.yaml")
	if fileErr != nil {
		fmt.Println("readConfig:", fileErr)
		// errors.Is 沿 Unwrap 链逐层比对，找到 os.ErrNotExist 就返回 true。
		// 比 fileErr == os.ErrNotExist 更好——后者在包装后就失配了。
		if errors.Is(fileErr, os.ErrNotExist) {
			fmt.Println("  → 文件不存在，请检查路径")
		}
	}

	// ── 5. 多层包装：业务 sentinel error 穿透实战 ──
	fmt.Println("\n=== 3. 多层包装穿透 ===")
	if err := runApp(); err != nil {
		fmt.Printf("完整错误链:\n%v\n", err) // %v 打印每层一行（Go 1.13+ 格式）
	}

	// ── 6. errors.Is vs errors.As — 一锯定音 ──
	// 一句话记忆：
	//   Is(链, 值) → 找「等不等于这个值」（sentinel）
	//   As(链, &指针) → 找「是不是这个类型」（提取字段用）
	fmt.Println("\n=== 4. Is vs As 对比 ===")
	err = fmt.Errorf("operation failed: %w", &ValidationError{Field: "age", Value: "-1", Msg: "负数"})
	fmt.Println("err:", err)
	fmt.Println("errors.Is(err, ErrNotFound):", errors.Is(err, ErrNotFound)) // false，不是同一个值
	var ve *ValidationError
	fmt.Println("errors.As(err, &ve):", errors.As(err, &ve)) // true，类型匹配
	if ve != nil {
		fmt.Printf("  → Field=%s, Value=%s, Msg=%s\n", ve.Field, ve.Value, ve.Msg)
	}

	// ── 7. errors.Join（Go 1.20+）：合并多个错误 ──
	fmt.Println("\n=== 5. errors.Join ===")
	e1 := errors.New("连接被拒绝")
	e2 := errors.New("认证过期")
	joined := errors.Join(e1, e2)
	fmt.Println("Join:", joined) // 每行一个错误
	// Is/As 对 joined 也有效——能匹配到里面任何一个。
	fmt.Println("Is e1:", errors.Is(joined, e1))
	fmt.Println("Is e2:", errors.Is(joined, e2))
}

// ── 实战函数 ──

// 业务 sentinel error：包级变量，命名以 Err 开头。
var (
	ErrDBConnection = errors.New("数据库连接失败")
	ErrTimeout      = errors.New("操作超时")
)

func readConfig(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// %w：把原始 err 嵌入新 error，形成链。
		// 新 error 的"表面消息"（%s/%v）是组合文本，但仍可通过 Unwrap 找到 os.ErrNotExist。
		return nil, fmt.Errorf("读取配置失败 %s: %w", path, err)
	}
	return data, nil
}

func queryUser(id int) error {
	// 模拟：ORM 层返回 sentinel error
	baseErr := ErrDBConnection
	// 用自定义 QueryError 包装，同时底层依旧能 Unwrap 到 sentinel
	return &QueryError{Query: fmt.Sprintf("SELECT * FROM users WHERE id=%d", id), Err: baseErr}
}

func runApp() error {
	err := queryUser(42)
	if err == nil {
		return nil
	}
	// errors.Is 层层 Unwrap，直到找到匹配的 sentinel。
	if errors.Is(err, ErrDBConnection) {
		fmt.Println("→ 检测到数据库故障，请检查连接")
	}
	// errors.As 判定自定义类型，拿到 Query 字段。
	var qe *QueryError
	if errors.As(err, &qe) {
		fmt.Println("→ QueryError.查询语句:", qe.Query)
	}
	return err
}

// ── 错误处理准则（代码即文档） ──

// 准则 1：调用方必须检查 err，不要忽略。
//   ✓  data, err := io.ReadAll(r); if err != nil { return err }
//   ✗  data, _ := io.ReadAll(r)  // 丢弃 error，炸了都不知道

// 准则 2：用 if err := ...; err != nil 限制 err 作用域。
//   ✓  if err := doWork(); err != nil { ... }
//   ✗  var err error; err = doWork(); if err != nil { ... } // err 逃逸到更大作用域

// 准则 3：包装时描述「做什么失败了」，底层错误用 %w 传入。
//   ✓  fmt.Errorf("查询用户(id=%d)失败: %w", id, err)
//   ✗  fmt.Errorf("用户错误: %s", err.Error()) // 丢失原始错误，Is/As 失效

// 准则 4：只在模块边界包装（repo → service → handler）。内部调用链不重复包装。
//   repo:  return fmt.Errorf("db: %w", err)  ← 边界处添加工位信息
//   service: if errors.Is(err, ErrDBConnection) { ... }  ← 只判断不包装

// 准则 5：sentinel error 命名以 Err 开头，语义明确。
//   ✓  ErrNotFound
//   ✗  ErrGeneric / ErrorFailed

// ── 练习 ──
// 1. 自定义 QueryError{Query string, Err error}，实现 Error() 和 Unwrap()，在 runApp 中通过 errors.As 提取 Query。
// 2. 用 fmt.Errorf 包装三次，分别用 errors.Is 和 errors.As 追溯。
// 3. 把 divide 改成返回 DivisionError{A, B float64}，实现 Error()。
