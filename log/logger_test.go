package log

import (
	"testing"
	"time"
	"fmt"
)

func TestNewLogger(t *testing.T) {
	NewLogger()
}

func TestLogger_Attach(t *testing.T) {

	logger := NewLogger()
	fileConfig := &FileConfig{
		Filename:"./test.log",
	}
	logger.Attach("file", LEVEL_DEBUG, fileConfig)
	outputs := logger.outputs
	for _, outputLogger := range outputs {
		if outputLogger.Name != "file" {
			t.Error("file attach failed")
		}
	}
}

func TestLogger_Detach(t *testing.T) {

	logger := NewLogger()
	logger.Detach("console")

	outputs := logger.outputs

	if len(outputs) > 0 {
		t.Error("logger detach error")
	}
}

func TestLogger_loggerMessageFormat(t *testing.T) {

	loggerMsg := &loggerMessage{
		Flag: "test",
		Nano : time.Now().UnixNano(),
		Time : time.Now().Format("2006-01-02 15:04:05.000"),
		Level : LEVEL_DEBUG,
		Type: "debug",
		Body: "logger console adapter test jsonFormat",
		File : "console_test.go",
		Line : 77,
		Func: "TestAdapterConsole_WriteJson",
	}

	format := "%millisecond_format% [%level_string%] [%file%:%line%] %body%"
	str := loggerMessageFormat(format, loggerMsg)

	fmt.Println(str)
}