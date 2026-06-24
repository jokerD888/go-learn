// === Lesson 07：interface — 隐式满足、类型断言、多态、组合 ===
// 目标：理解 Go 的隐式接口机制——无需声明 implements，方法集匹配即满足。

package main

import (
	"fmt"
	"math"
)

// ── 1. interface 声明：只定义行为，不定义数据 ──
// Go 的 interface = 方法集合。任何类型只要实现了接口的全部方法，
// 就隐式满足该接口——不需要像 Java 那样写 "implements Speaker"。
// 这叫"结构化类型"（structural typing）：只看方法签名是否匹配，不看类型名字。
//
// 对比：
//   Java:  class Dog implements Speaker { ... }     ← 编译时绑定
//   Go:    type Dog struct{}; func (d Dog) Speak() … ← 写完方法自动满足

type Speaker interface {
	Speak() string
}

// Dog 有 Speak() → 自动满足 Speaker。无需任何声明。
type Dog struct{ Name string }

func (d Dog) Speak() string { return fmt.Sprintf("汪！我是 %s", d.Name) }

// Cat 也有 Speak() → 也满足 Speaker。一个类型可同时满足多个接口。
type Cat struct{ Name string }

func (c Cat) Speak() string { return fmt.Sprintf("喵——我是 %s", c.Name) }

// ── 2. 多态：接口变量可持有任意满足该接口的值 ──
// 运行时接口变量存的是 (concrete_type, concrete_value) 对，调方法时做动态分发。

func main() {
	fmt.Println("=== 2. 多态 ===")
	var s Speaker        // 零值是 nil，没有具体类型——调 Speak() 会 panic
	s = Dog{Name: "旺财"} // s = (Dog, Dog{Name:"旺财"})
	fmt.Println(s.Speak())
	s = Cat{Name: "咪咪"} // s = (Cat, Cat{Name:"咪咪"})
	fmt.Println(s.Speak())

	// ── 3. 接口作为函数参数：依赖倒置的基石 ──
	// announce 不依赖 Dog/Cat 的具体类型，只依赖 Speaker 接口。
	// 新增任何实现了 Speak() 的类型都无需改 announce 的代码。
	fmt.Println("\n=== 3. 接口参数 ===")
	announce(Dog{Name: "大黄"})
	announce(Cat{Name: "小白"})

	// ── 4. 空接口 any — 万能容器 ──
	// any = interface{}（Go 1.18 起 any 是 alias）。没有任何方法 → 所有类型都满足。
	// 可以把它想象成"无约束的类型参数"，但存入的值要取回必须类型断言。
	fmt.Println("\n=== 4. any ===")
	var anything any
	anything = 42
	fmt.Println("any 存 int:", anything)
	anything = "hello"
	fmt.Println("any 存 string:", anything)
	anything = Dog{Name: "小黑"}
	fmt.Println("any 存 Dog:", anything)

	// ── 5. 类型断言：从接口恢复具体类型 ──
	// 接口把具体类型"装箱"了，断言把它拆出来。
	// 两种写法：
	//   v := x.(T)        // 失败直接 panic
	//   v, ok := x.(T)    // 安全：ok=false 时 v 是 T 的零值
	fmt.Println("\n=== 5. 类型断言 ===")
	printIfDog(Dog{Name: "阿福"})
	printIfDog(Cat{Name: "汤圆"})

	// ── 6. 类型 switch：按类型分支 ──
	// 只能用在 switch 中，语法是 x.(type)。内部 case 匹配后 v 自动变成对应类型。
	fmt.Println("\n=== 6. 类型 switch ===")
	describe("一串文字")
	describe(3.14)
	describe(42)
	describe(Dog{Name: "大毛"})

	// ── 7. 指针接收者 vs 值接收者：谁满足接口？ ──
	// 方法接收者是 *Document（指针），只有 *Document 才实现了 Read/Write。
	// Document 值类型没有这两个方法 → 不满足 Reader/Writer。
	// 规则：
	//   值接收者方法 → 值和指针都满足接口
	//   指针接收者方法 → 只有指针满足接口（值类型不能取地址传给指针参数）
	fmt.Println("\n=== 7. 接收者与接口 ===")
	var doc = &Document{title: "go101", content: "interface 是 Go 的灵魂"}
	var r Reader = doc  // ✓ *Document 满足 Reader
	// var r2 Reader = Document{...}  // ✗ 编译错：Document 没有 Read 方法（Read 定义在 *Document 上）
	fmt.Println("读取前 5 字节:", readFirst5(r))

	// ── 8. 接口组合：小接口拼大接口 ──
	// Reader + Writer → ReadWriter。面向对象靠继承，Go 靠组合。
	// 标准库准则：定义 1-3 个方法的小接口，用时组合。io.Reader 只有 1 个方法。
	fmt.Println("\n=== 8. 接口组合 ===")
	var rw ReadWriter = doc // *Document 有 Read+Write → 满足 ReadWriter
	rw.Write("hello")
	fmt.Println("written:", doc.written)

	// ── 9. 空接口性能 ──
	// interface 底层是 iface（有方法接口）或 eface（空接口），
	// 都是一个 (type 指针, data 指针) 对。每次调方法需查 itab → 有间接跳转开销。
	// 热路径传具体类型，冷路径用接口做抽象——Go 性能直觉。
}

// ── 辅助函数 ──

func announce(s Speaker) {
	fmt.Println("📢", s.Speak())
}

// 类型断言安全写法：ok 模式避免 panic。
func printIfDog(s Speaker) {
	if dog, ok := s.(Dog); ok {
		fmt.Printf("是狗！%s\n", dog.Name)
	} else {
		fmt.Println("不是狗")
	}
}

// 类型 switch 中 v 在每个 case 里自动转型为对应具体类型。
func describe(i any) {
	switch v := i.(type) {
	case string:
		fmt.Printf("字符串: %q, 长度=%d\n", v, len(v))
	case float64:
		fmt.Printf("浮点数，开方: %.2f\n", math.Sqrt(v))
	case int:
		fmt.Printf("整数: %d\n", v)
	case Dog:
		fmt.Printf("狗: %s\n", v.Name)
	default:
		fmt.Printf("未知类型 %T: %v\n", v, v)
	}
}

// ── 9. fmt.Stringer — 最常用的标准库接口 ──
// 实现 String() string 就能控制 fmt.Println / fmt.Printf("%s") 的输出。
// 类似于 Java 的 toString()、Python 的 __str__。

type User struct {
	Name  string
	Email string
}

func (u User) String() string {
	return fmt.Sprintf("User(%s, %s)", u.Name, u.Email)
}

// ── 10. 接口组合 — Go 的抽象基石 ──

type Reader interface {
	Read(n int) string
}

type Writer interface {
	Write(data string)
}

// 接口内嵌接口：ReadWriter 要求同时实现 Read 和 Write。
// 和 struct 内嵌一样，Go 用组合而非继承来扩展能力。
type ReadWriter interface {
	Reader // 展开 = Read(n int) string
	Writer // 展开 = Write(data string)
}

type Document struct {
	title   string
	content string
	written string
}

// 注意：Read/Write 定义在 *Document 上（指针接收者）。
// 只有 *Document 满足 Reader/Writer/ReadWriter，Document 本身不满足。
func (d *Document) Read(n int) string {
	if n > len(d.content) {
		n = len(d.content)
	}
	return d.content[:n]
}

func (d *Document) Write(data string) {
	d.written += data
}

// "接受最小接口，返回具体类型"——Go 设计准则。
// r 只要 Reader（最小接口），调用方无论传 *Document / *bytes.Buffer / *os.File 都行。
func readFirst5(r Reader) string {
	return r.Read(5)
}

// ── 练习 ──
// 1. 定义 Greeter 接口（Greet() string），给 Dog/Cat 实现它。
// 2. 类型 switch：printType(any) 对 int/string/bool 分类输出。
// 3. 给 User 实现 String()，输出 "User(name=..., email=...)"。
// 4. 定义一个 PointerSpeaker 接口，要求方法接收者是指针。把 Dog 的 Speak 改成指针接收者，看值类型还能不能赋值给接口。
