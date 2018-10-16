package log

import (
	"testing"
	"time"
)

func TestNewAdapterFile(t *testing.T) {
	NewAdapterFile()
}

func TestAdapterFile_Name(t *testing.T) {
	fileAdapter := NewAdapterFile()

	if fileAdapter.Name() != FILE_ADAPTER_NAME {
		t.Error("file adapter name error")
	}
}

func TestAdapterFile_Write(t *testing.T) {

	fileAdapter := NewAdapterFile()

	fileConfig := &FileConfig{
		Filename: "./test.log",
		LevelFileName: map[int]string{

		},
		MaxLine:2000,
		MaxSize: 10000*4,
		Json:true,
		DateSlice:"d",
	}
	err := fileAdapter.Init(fileConfig)
	if err != nil {
		t.Error(err.Error())
	}

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
	err = fileAdapter.Write(loggerMsg)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestAdapterFile_WriteLevelFile(t *testing.T) {

	fileAdapter := NewAdapterFile()

	fileConfig := &FileConfig{
		Filename: "./test.log",
		LevelFileName: map[int]string{
			LEVEL_DEBUG: "./debug.log",
			LEVEL_INFO: "./info.log",
			LEVEL_ERROR: "./error.log",
		},
		MaxLine:2000,
		MaxSize: 10000*4,
		Json:true,
		DateSlice:"d",
	}
	err := fileAdapter.Init(fileConfig)
	if err != nil {
		t.Error(err.Error())
	}

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
	fileAdapter.Write(loggerMsg)
	loggerMsg.Level = LEVEL_INFO
	fileAdapter.Write(loggerMsg)
	loggerMsg.Level = LEVEL_ERROR
	fileAdapter.Write(loggerMsg)
}