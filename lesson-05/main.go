// === Lesson 05：复合类型 — slice/map/struct、make vs new ===
// 目标：掌握 Go 的三种复合类型，理解值类型 vs 引用语义。

package main

import "fmt"

func main() {
	// ── array：定长，长度是类型的一部分 ──
	// [3]int 和 [4]int 是不同的类型。array 几乎不用——用 slice。
	var arr [3]int
	arr[0] = 1
	arr[1] = 2
	arr[2] = 3
	fmt.Println("array:", arr)

	// ── slice：动态数组，Go 里 99% 的"数组"都是 slice ──
	// slice 是一个 header（ptr + len + cap），底层指向一个 array。
	// 这是 slice 和 C++ vector 的本质区别：slice 只是"视图"，多 slice 可共享底层 array。

	// 创建方式 1：字面量
	s1 := []int{10, 20, 30}
	fmt.Println("slice 字面量:", s1, "len:", len(s1), "cap:", cap(s1))

	// 创建方式 2：make —— 预分配容量，避免 append 过程中的反复扩容。
	// make 只用于 slice/map/chan，new 几乎不用。
	s2 := make([]int, 3, 5) // len=3, cap=5
	fmt.Println("make slice:", s2, "len:", len(s2), "cap:", cap(s2))

	// append：追加元素。cap 不够时自动扩容（通常是 2x 增长）。
	s2 = append(s2, 40)
	fmt.Println("append 后:", s2, "len:", len(s2), "cap:", cap(s2))

	// ── 切片操作：[low:high] 返回新 slice header，底层 array 共享 ──
	// 这是性能关键：切片不复制数据，只复制 header。
	original := []int{0, 1, 2, 3, 4, 5}
	view := original[1:4] // [1,2,3]，共享底层 array
	fmt.Println("切片前:", original, "切片:", view)
	view[0] = 999          // 修改 view，original 也变了——共享底层
	fmt.Println("修改后:", original, "切片:", view)

	// ── map：无序键值对，nil map 写会 panic ──
	// map 底层是哈希表，key 必须能用 == 比较（slice 不能当 key）。
	scores := map[string]int{
		"alice": 95,
		"bob":   87,
	}
	scores["charlie"] = 92 // 写
	delete(scores, "bob")  // 删
	fmt.Println("map:", scores)

	// 安全的读取：用 ok 检查 key 是否存在。
	// 不要依赖零值判断——key 不存在返回零值 0，但 value 本身可能是 0。
	if _, ok := scores["bob"]; !ok {
		fmt.Println("bob 不存在")
	}

	// for range map 遍历顺序随机，Go 故意打乱——防止你依赖顺序。
	for k, v := range scores {
		fmt.Printf("  %s: %d\n", k, v)
	}

	// ── struct：字段集合，Go 没 class，struct 就是"对象" ──
	type Point struct {
		X, Y int // 大写 = 公开，小写 = 包内私有
	}
	p1 := Point{10, 20}            // 按位置初始化
	p2 := Point{X: 5, Y: 15}       // 按字段名初始化（推荐）
	p2.X = 100                      // . 访问字段
	fmt.Printf("p1=%+v p2=%+v\n", p1, p2)

	// ── new vs make ──
	// new(T) 分配堆内存，返回 *T —— 极少用。
	// make(T, ...) 只用于 slice/map/chan，返回 T（不是指针）。
	// ponytail: new 几乎只在实现时出现，日常写代码全用 make 和 &T{}。
	p3 := new(Point)   // *Point
	p3.X = 1           // 指针自动解引用（语法糖，等于 (*p3).X）
	fmt.Printf("new(Point) = %+v\n", *p3)
}

// ── 练习 ──
// 1. 对 []int{1,2,3,4,5} 做切片 [1:4]，修改一个元素，验证底层共享。
// 2. 遍历 map 三次，观察顺序是否每次都不同。
// 3. 声明 var m map[string]int，不给它 make，直接写 m["x"]=1，看 panic。
