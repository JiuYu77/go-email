package smtp

import (
	"fmt"
)

func addr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
