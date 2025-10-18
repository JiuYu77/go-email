package logx

import "time"

const (
	DateTime = time.DateTime + ".000"
)

func now_str(layout string) string {
	if layout == "" {
		layout = DateTime
	}
	return time.Now().Format(layout)
}
