package logx

import "fmt"

type Style uint8

const (
	// 重置
	Reset Style = iota

	// 字体样式 font style 、misc
	Bold
	HalfBright // 低亮
	Italic
	Underline
	Blink         // 闪烁（无效）
	_             // 占位，多余的一个，没有6
	Reverse       // 反显
	Blanking      // 消隐
	StrikeThrough // 删除线
)

// 字体颜色 basic color
const (
	Black Style = iota + 30 // 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// 字体颜色 亮色 bright color
const (
	BtBlack Style = iota + 90 // 90
	BtRed
	BtGreen
	BtYellow
	BtBlue
	BtMagenta
	BtCyan
	BtWhite
)

// 背景颜色 background color
const (
	BgBlack Style = iota + 40 // 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// 将单个样式添加到字符串
func (s Style) Add(str string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", s, str)
}

// 为字符串一次添加多个样式
// opts {...Style}
func AddStyle(s string, opts ...Style) string {
	if len(opts) == 0 {
		return s
	}
	var res string
	for _, v := range opts {
		res = fmt.Sprintf("%s\x1b[%dm", res, v)
	}
	res = res + s + "\x1b[0m"
	return res
}
