package log

import (
	"os"
	"runtime"
)

func PrintPanicInfo(format string, err any) {
	// 错误业务日志
	Errorf(format, err)

	// 获取并打印堆栈信息
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	Errorf("Stack trace:\\n%s", buf[:n])
	os.Exit(-1)
}
