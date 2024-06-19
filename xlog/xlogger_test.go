package xlog

import (
	"fmt"
	"log"
	"os"
	"testing"
)

var mockLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func TestLogWithPosition(t *testing.T) {
	fmt.Println("1111111")
	logWithPosition(mockLogger, "这是一条日志信息", "32")
}
