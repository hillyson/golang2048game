package g2048

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

var Score int
var step int

// 输出字符串
func coverPrintStr(x, y int, str string, fg, bg termbox.Attribute) error {

	xx := x
	for n, c := range str {
		if c == '\n' {
			y++
			xx = x - n - 1
		}
		termbox.SetCell(xx+n, y, c, fg, bg)
	}
	_ = termbox.Flush()
	return nil
}

// 游戏状态
type Status uint

const (
	Win Status = iota
	Lose
	Add
	Max = 2048
)

// 2048游戏中的16个格子使用4x4二维数组表示
type G2048 [4][4]int

// 检查游戏是否已经胜利，没有胜利的情况下随机将值为0的元素
// 随机设置为2或者4
func (m *G2048) checkWinOrAdd() Status {
	// 判断4x4中是否有元素的值大于(等于)2048，有则获胜利
	for _, row := range m {
		for _, cell := range row {
			if cell >= Max {
				return Win
			}
		}
	}
	// 开始随机设置零值元素为2或者4
	i := rand.Intn(len(m))
	j := rand.Intn(len(m))
	for x := 0; x < len(m); x++ {
		for y := 0; y < len(m); y++ {
			if m[i%len(m)][j%len(m)] == 0 {
				//通过随机移位0或1的方式决定数字是2还是2*2(4)
				m[i%len(m)][j%len(m)] = 2 << (rand.Uint32() % 2)
				return Add
			}
			j++
		}
		i++
	}

	// 全部元素都不为零（表示已满），则失败
	return Lose
}

// 初始化游戏界面
func (m G2048) initialize(ox, oy int) error {
	fg := termbox.ColorYellow
	bg := termbox.ColorBlack
	_ = termbox.Clear(fg, bg)
	str := "     分数: " + fmt.Sprint(Score)
	for n, c := range str {
		termbox.SetCell(ox+n, oy-1, c, fg, bg)
	}
	str = "退出:ESC " + "重来:ENTER"
	for n, c := range str {
		termbox.SetCell(ox+n, oy-2, c, fg, bg)
	}
	str = "通过方向键移动"
	for n, c := range str {
		termbox.SetCell(ox+n, oy-3, c, fg, bg)
	}

	ox -= 1
	fg = termbox.ColorBlack
	bg = termbox.ColorGreen
	for i := 0; i <= len(m); i++ {
		for x := 0; x < 6*len(m)+1; x++ {
			if x%6 != 0 {
				termbox.SetCell(ox+x, oy+i*2, '-', fg, bg)
			}
		}
		for y := 0; y <= 2*len(m); y++ {
			if y%2 == 0 {
				termbox.SetCell(ox+i*6, oy+y, '+', fg, bg)
			} else {
				termbox.SetCell(ox+i*6, oy+y, '|', fg, bg)
			}
		}
	}
	fg = termbox.ColorRed
	bg = termbox.ColorBlack
	for i := range m {
		for j := range m[i] {
			if m[i][j] > 0 {
				str := fmt.Sprint(m[i][j])
				for n, char := range str {
					termbox.SetCell(ox+j*6+3+n-(len(str)/2), oy+i*2+1, char, fg, bg)
				}
			}
		}
	}
	return termbox.Flush()
}

// 翻转二维切片
func (m *G2048) mirrorV() {
	tn := new(G2048)
	for i, row := range m {
		for j, cell := range row {
			tn[len(m)-i-1][j] = cell
		}
	}
	*m = *tn
}

// 向右旋转90度
func (m *G2048) right90() {
	tn := new(G2048)
	for i, row := range m {
		for j, cell := range row {
			tn[j][len(m)-i-1] = cell
		}
	}
	*m = *tn
}

// 向左旋转90度
func (m *G2048) left90() {
	tn := new(G2048)
	for i, row := range m {
		for j, cell := range row {
			tn[len(row)-j-1][i] = cell
		}
	}
	*m = *tn
}

// 旋转180度
func (m *G2048) right180() {
	tn := new(G2048)
	for i, row := range m {
		for j, cell := range row {
			tn[len(row)-i-1][len(row)-j-1] = cell
		}
	}
	*m = *tn
}

// 向上移动并合并
func (m *G2048) mergeUp() bool {
	maxWidth := len(m)
	changed := false
	notFull := false
	for i := 0; i < maxWidth; i++ {

		maxCheck := maxWidth
		index := 0 // 统计每一列中非零值的个数

		// 向上移动非零值，如果有零值元素，则用非零元素进行覆盖
		for x := 0; x < maxWidth; x++ {
			if m[x][i] != 0 {
				m[index][i] = m[x][i]
				if index != x {
					changed = true // 标示数组的元素是否有变化
				}
				index++
			}
		}
		if index < maxWidth {
			notFull = true
		}
		maxCheck = index
		// 向上合并所有相同的元素
		for x := 0; x < maxCheck-1; x++ {
			if m[x][i] == m[x+1][i] {
				m[x][i] *= 2
				m[x+1][i] = 0
				Score += m[x][i] * step // 计算游戏分数
				x++
				changed = true
			}
		}
		// 合并完相同元素以后，再次向上移动非零元素
		index = 0
		for x := 0; x < maxCheck; x++ {
			if m[x][i] != 0 {
				m[index][i] = m[x][i]
				index++
			}
		}
		// 对于没有检查的元素
		for x := index; x < maxWidth; x++ {
			m[x][i] = 0
		}
	}
	return changed || !notFull
}

// 向下移动合并的操作可以转换向上移动合并:
// 1. 翻转切片
// 2. 向上合并
// 3. 再次翻转切片，得到原切片向下合并的结果
func (m *G2048) mergeDown() bool {
	//t.mirrorV()
	m.right180()
	changed := m.mergeUp()
	//t.mirrorV()
	m.right180()
	return changed
}

// 向左移动合并转换为向上移动合并
func (m *G2048) mergeLeft() bool {
	m.right90()
	changed := m.mergeUp()
	m.left90()
	return changed
}

/// 向右移动合并转换为向上移动合并
func (m *G2048) mergeRight() bool {
	m.left90()
	changed := m.mergeUp()
	m.right90()
	return changed
}

// 检查按键，做出不同的移动动作或者退出程序
func (m *G2048) mergeAndReturnKey(eventQueue chan termbox.Event) termbox.Key {
	var changed bool
	changed = false
	//ev := termbox.PollEvent()
Label:
	ev := <-eventQueue
	//fmt.Printf("%s%s", "感受到按键", time.Now())
	switch ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeyArrowUp:
			changed = m.mergeUp()
		case termbox.KeyArrowDown:
			changed = m.mergeDown()
		case termbox.KeyArrowLeft:
			changed = m.mergeLeft()
		case termbox.KeyArrowRight:
			changed = m.mergeRight()
		case termbox.KeyEsc, termbox.KeyEnter:
			changed = true
		default:
			changed = false
		}

		// 如果元素的值没有任何更改，则从新开始循环
		if !changed {
			goto Label
		}

	case termbox.EventResize:
		x, y := termbox.Size()
		_ = m.initialize(x/2-10, y/2-4)
		goto Label
	case termbox.EventError:
		panic(ev.Err)
	}
	step++ // 计算游戏操作数
	return ev.Key
}

// 重置
func (m *G2048) clear() {
	next := new(G2048)
	Score = 0
	step = 0
	*m = *next

}

// 开始游戏
func (m *G2048) Run() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())

A:

	m.clear()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent() // 开始监听键盘事件
		}
	}()

	for {
		gameResult := m.checkWinOrAdd()
		x, y := termbox.Size()
		_ = m.initialize(x/2-10, y/2-4)
		switch gameResult {
		case Win:
			str := "Win!!"
			strLength := len(str)
			_ = coverPrintStr(x/2-strLength/2, y/2, str, termbox.ColorMagenta, termbox.ColorYellow)
		case Lose:
			str := "Lose!!"
			strLength := len(str)
			_ = coverPrintStr(x/2-strLength/2, y/2, str, termbox.ColorBlack, termbox.ColorRed)
		case Add:
			_ = m.initialize(x/2-10, y/2-4)
		default:
			fmt.Print("Err")
		}
		// 检查用户按键
		key := m.mergeAndReturnKey(eventQueue)
		// 如果按键是 Esc 则退出游戏
		if key == termbox.KeyEsc {
			return
		}
		// 如果按键是 Enter 则从新开始游戏
		if key == termbox.KeyEnter {
			goto A
		}
	}
}
