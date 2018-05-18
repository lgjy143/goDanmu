package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// log level
const (
	LevelFatal = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
)

// log output type
const (
	ConsoleLog = iota
	FileLog
)

type LevelLog struct {
	Logger *log.Logger
	Level  int
	Type   int
	Output io.Writer
}

var Log LevelLog
var LogFormat int = log.Ldate | log.Ltime | log.Lshortfile

func init() {
	Log.Logger = log.New(os.Stderr, "", LogFormat)
	Log.Level = LevelWarning
	Log.Type = ConsoleLog
	Log.Output = os.Stderr
}

func SetType(logType int, conf ...map[string]string) {
	switch logType {
	case FileLog:
		fileName := conf[0]["fileName"]
		errorLog, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
			os.Exit(1)
		}
		Log.Logger = log.New(errorLog, "", LogFormat)
		Log.Type = FileLog
		Log.Output = errorLog
	case ConsoleLog:
		break
	default:
		break
	}
}

func SetLevel(logLevel int) error {
	if logLevel >= LevelFatal && logLevel <= LevelDebug {
		Log.Level = logLevel
		return nil
	}
	return errors.New("error input log level")
}

func Info(v ...interface{}) {
	if Log.Level >= LevelInfo {
		Log.Logger.SetPrefix("[Info] ")
		Log.Logger.Print(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if Log.Level >= LevelInfo {
		Log.Logger.SetPrefix("[Info] ")
		Log.Logger.Printf(format, v...)
	}
}

func Warning(v ...interface{}) {
	if Log.Level >= LevelWarning {
		Log.Logger.SetPrefix("[Warning] ")
		Log.Logger.Print(v...)
	}
}

func Warningf(format string, v ...interface{}) {
	if Log.Level >= LevelWarning {
		Log.Logger.SetPrefix("[Warning] ")
		Log.Logger.Printf(format, v...)
	}
}

func Error(v ...interface{}) {
	if Log.Level >= LevelError {
		Log.Logger.SetPrefix("[Error] ")
		Log.Logger.Print(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if Log.Level >= LevelError {
		Log.Logger.SetPrefix("[Error] ")
		Log.Logger.Printf(format, v...)
	}
}

func Fatal(v ...interface{}) {
	if Log.Level >= LevelFatal {
		Log.Logger.SetPrefix("[Fatal] ")
		Log.Logger.Fatal(v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if Log.Level >= LevelFatal {
		Log.Logger.SetPrefix("[Fatal] ")
		Log.Logger.Fatalf(format, v...)
	}
}
