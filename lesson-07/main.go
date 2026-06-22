// === Lesson 07：interface — 隐式满足、类型断言、多态 ===
// 目标：理解 Go 的隐式接口——Go 最强大的抽象机制，没有之一。

package main

import (
	"fmt"
	"math"
)

// ── interface 声明：只定义行为，不定义数据 ──
// Go 的 interface 是"鸭子类型"的静态版本：只要你叫起来像鸭子，你就是鸭子。
// 不需要显式声明 "implements"——只要方法签名匹配就算实现。

type Speaker interface {
	Speak() string
}

// Dog 有 Speak() 方法 → 自动实现了 Speaker 接口。无需写 implements。
type Dog struct{ Name string }

func (d Dog) Speak() string { return fmt.Sprintf("汪！我是 %s", d.Name) }

// Cat 也有 Speak() → 也实现了 Speaker。一个类型可以实现多个接口。
type Cat struct{ Name string }

func (c Cat) Speak() string { return fmt.Sprintf("喵——我是 %s", c.Name) }

func main() {
	// ── 多态：同一个接口变量可以持有不同类型的值 ──
	var s Speaker
	s = Dog{Name: "旺财"}
	fmt.Println(s.Speak())
	s = Cat{Name: "咪咪"}
	fmt.Println(s.Speak())

	// ── 接口作为函数参数：接受任何实现了接口的类型 ──
	announce(Dog{Name: "大黄"})
	announce(Cat{Name: "小白"})

	// ── 空接口 any（Go 1.18+，老代码用 interface{}） ──
	// any 没有任何方法 → 所有类型都实现了 any。
	// 相当于 Java 的 Object / C 的 void*，但有类型安全。
	var anything any
	anything = 42
	fmt.Println("any 存 int:", anything)
	anything = "hello"
	fmt.Println("any 存 string:", anything)
	anything = Dog{Name: "小黑"}
	fmt.Println("any 存 Dog:", anything)

	// ── 类型断言：从接口中取回具体类型 ──
	// 语法：x.(T)，x 是接口变量，T 是具体类型。
	printIfDog(Dog{Name: "阿福"})
	printIfDog(Cat{Name: "汤圆"})

	// ── 类型 switch：一次性判断多种类型 ──
	describe("一串文字")
	describe(3.14)
	describe(Dog{Name: "大毛"})

	// ── 标准库接口最佳实践 ──
	// fmt.Stringer = 自定义打印格式。实现了它就控制了 fmt.Println 的输出。
	user := User{Name: "alice", Email: "alice@example.com"}
	fmt.Println(user) // 调用了 String()

	// ── 接口组合：大接口由小接口拼成 ──
	// Go 鼓励定义小接口（1-3 个方法），用组合拼成大接口。
	// io.Reader、io.Writer 等都只有 1 个方法。
	var doc Document = Document{title: "go101", content: "interface 是 Go 的灵魂"}
	fmt.Println("读取前 5 字节:", readFirst5(&doc))

	// ── 空接口 vs 具体类型：性能代价 ──
	// interface 变量底层是一个 (type, value) 对，每次调用方法都有一次间接跳转。
	// ponytail: 频繁调用的热路径不用 interface，性能敏感的地方传具体类型。
}

// announce 接受 Speaker 接口，不知道也不关心是狗还是猫。
func announce(s Speaker) {
	fmt.Println("📢", s.Speak())
}

// 类型断言：.(type) 检查是否是某种具体类型。
func printIfDog(s Speaker) {
	if dog, ok := s.(Dog); ok { // ok=true 表示断言成功
		fmt.Printf("是狗！名字叫 %s\n", dog.Name)
	} else {
		fmt.Println("不是狗")
	}
}

// 类型 switch：比 if-断言链更清晰。
func describe(i any) {
	switch v := i.(type) { // .(type) 只能在 switch 语句中使用
	case string:
		fmt.Printf("字符串，长度: %d\n", len(v))
	case float64:
		fmt.Printf("浮点数，开方: %.2f\n", math.Sqrt(v))
	case Dog:
		fmt.Printf("狗，名字: %s\n", v.Name)
	default:
		fmt.Printf("未知类型: %T\n", v)
	}
}

// ── fmt.Stringer：实现 String() string 即可自定义打印 ──
type User struct {
	Name  string
	Email string
}

func (u User) String() string {
	return fmt.Sprintf("User(%s, %s)", u.Name, u.Email)
}

// ── 接口组合 ──

type Reader interface {
	Read(n int) string
}

type Writer interface {
	Write(data string)
}

// ReadWriter = Reader + Writer，用接口内嵌接口来组合。
type ReadWriter interface {
	Reader
	Writer
}

type Document struct {
	title   string
	content string
	written string
}

func (d *Document) Read(n int) string {
	if n > len(d.content) {
		n = len(d.content)
	}
	return d.content[:n]
}

func (d *Document) Write(data string) {
	d.written += data
}

// 接受最小的接口，返回具体的类型——Go 设计原则。
func readFirst5(r Reader) string {
	return r.Read(5)
}

// ── 练习 ──
// 1. 定义一个 Greeter 接口（Greet() string），让 Dog 和 Cat 都实现它。
// 2. 用类型 switch 写一个 printType 函数，对 int/string/bool 分别输出。
// 3. 给 User 实现 String() 方法，打印 "User(name: ..., email: ...)"。
