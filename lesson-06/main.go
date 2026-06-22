// === Lesson 06：方法与指针 — 值接收者 vs 指针接收者 ===
// 目标：理解 Go 的方法机制，知道何时用值接收者、何时用指针接收者。

package main

import "fmt"

// ── 方法 = 带有接收者的函数 ──
// Go 没有 class。给任意类型（必须在本包内定义）绑定方法就能当"对象"用。

type Counter struct {
	value int
}

// 值接收者：c Counter —— 操作的是副本，原值不变。
// 适用场景：方法不需要修改 struct、struct 很小（≤几个字段）、并发安全。
func (c Counter) Value() int {
	return c.value
}

// 指针接收者：c *Counter —— 操作的是原值，可以修改。
// 适用场景：需要修改 struct、struct 很大（避免复制开销）、必须用指针的场景。
func (c *Counter) Inc() {
	c.value++ // 等价于 (*c).value++，Go 自动解引用
}

func (c *Counter) Reset() {
	c.value = 0
}

func main() {
	c := Counter{value: 10}
	fmt.Println("初始:", c.Value())

	c.Inc() // c 自动取地址——Go 的语法糖，等于 (&c).Inc()
	c.Inc()
	fmt.Println("Inc 两次:", c.Value())

	c.Reset()
	fmt.Println("Reset 后:", c.Value())

	// ── 指针接收者的关键场景 ──
	acc := BankAccount{
		owner:   "alice",
		balance: 1000,
	}
	fmt.Printf("\n账户: %s, 余额: %d\n", acc.owner, acc.balance)

	// 用指针接收者方法修改余额。
	acc.Deposit(500)
	fmt.Println("存入 500:", acc.balance)
	if err := acc.Withdraw(200); err != nil {
		fmt.Println(err)
	}
	fmt.Println("取出 200:", acc.balance)
	// 余额不足——返回 error，不会修改 balance。
	if err := acc.Withdraw(9999); err != nil {
		fmt.Println("取款失败:", err)
	}

	// ── 方法可以绑定到任何本包类型 ──
	// 包括基础类型的别名。Go 的类型系统=鸭子+结构。
	temps := Temperatures{23, 18, 30, 15, 27}
	fmt.Printf("\n温度: %v, 平均: %.1f\n", temps, temps.Avg())

	// ── 值接收者 vs 指针接收者：选择规则 ──
	// 规则很简单：
	//   需要修改 → 指针接收者
	//   不需要修改 + 小 struct → 值接收者
	//   不需要修改 + 大 struct → 指针接收者（省复制）
	//   必须实现某 interface → 看 interface 定义的接收者类型
	// 一个类型的方法应该统一——要么全值、要么全指针（混用只在少数场景合理）。
}

// ── 完整例子：银行账户 ──

type BankAccount struct {
	owner   string
	balance int
}

// 存款——要修改余额，指针接收者。
func (a *BankAccount) Deposit(amount int) {
	a.balance += amount
}

// 取款——要修改余额 + 返回可能的错误。
func (a *BankAccount) Withdraw(amount int) error {
	if amount > a.balance {
		return fmt.Errorf("余额不足: 需要 %d, 当前 %d", amount, a.balance)
	}
	a.balance -= amount
	return nil
}

// ── 方法绑定到切片类型 ──

type Temperatures []int

// Avg 只读操作，但 slice 本身是 header（24 字节），值接收者完全够用。
// 值接收者：调用时传的是 header 拷贝，但底层 array 共享——遍历效率一样。
func (t Temperatures) Avg() float64 {
	if len(t) == 0 {
		return 0
	}
	sum := 0
	for _, v := range t {
		sum += v
	}
	return float64(sum) / float64(len(t))
}

// ── 练习 ──
// 1. 给 Counter 加一个 Add(n int) 方法，一次加 n。
// 2. 把 Inc 改成值接收者，看 c.Inc() 后 Value() 是否变化。
// 3. 给 BankAccount 加 Transfer(to *BankAccount, amount int) error 方法。
