package log

import (
	"testing"
	"time"
)

func TestNewAdapterConsole(t *testing.T) {
	NewAdapterConsole()
}

func TestAdapterConsole_Init(t *testing.T) {

	consoleAdapter := NewAdapterConsole()

	consoleConfig := &ConsoleConfig{}
	consoleAdapter.Init(consoleConfig)
}

func TestAdapterConsole_Name(t *testing.T) {

	consoleAdapter := NewAdapterConsole()

	if consoleAdapter.Name() != CONSOLE_ADAPTER_NAME {
		t.Error("consoleAdapter name error")
	}
}


func TestAdapterConsole_WriteJson(t *testing.T) {

	consoleAdapter := NewAdapterConsole()

	consoleConfig := &ConsoleConfig{
		Json:true,
	}
	consoleAdapter.Init(consoleConfig)

	loggerMsg := &loggerMessage {
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
	err := consoleAdapter.Write(loggerMsg)
	if err != nil {
		t.Error(err.Error())
	}
}