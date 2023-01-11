package log

import (
	"fmt"

	"github.com/fatedier/beego/logs"
)

var Log *logs.BeeLogger

func init() {
	Log = logs.NewLogger(200)
	Log.EnableFuncCallDepth(true)
	Log.SetLogFuncCallDepth(Log.GetLogFuncCallDepth() + 1)
}

func InitLog(logWay string, logFile string, logLevel string, disableLogColor bool) {
	SetLogFile(logWay, logFile, disableLogColor)
	SetLogLevel(logLevel)
}

func SetLogFile(logWay string, logFile string, disableLogColor bool) {
	var params string
	switch logWay {
	case "file":
		// file
		params = fmt.Sprintf(`{"filename": "%s"}`, logFile)
		_ = Log.SetLogger("file", params)
		return
	case "console":
		goto console
	default:
		// other
		Log.Warn("log_way error... set log_way is \"console\".")
		goto console
	}
console:
	// defaule console
	if disableLogColor {
		params = `{"color": false}`
	}
	_ = Log.SetLogger("console", params)
}

func SetLogLevel(logLevel string) {
	var level int
	switch logLevel {
	case "error":
		level = 3
	case "warn":
		level = 4
	case "info":
		level = 6
	case "debug":
		level = 7
	case "trace":
		level = 8
	default:
		Log.Warn("log_level error... set log_level is \"info\".")
		level = 6
	}
	Log.SetLevel(level)
}

func Error(format string, v ...interface{}) {
	Log.Error(format, v...)
}

func Warn(format string, v ...interface{}) {
	Log.Warn(format, v...)
}

func Info(format string, v ...interface{}) {
	Log.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	Log.Debug(format, v...)
}

func Trace(format string, v ...interface{}) {
	Log.Trace(format, v...)
}
