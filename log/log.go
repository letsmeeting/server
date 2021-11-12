package log

import (
	"fmt"
	"github.com/jinuopti/lilpop-server/configure"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}
	return sprint
}

type logger struct {
	fpLog *os.File

	disableDebug bool
	disableInfo  bool
	disableError bool

	logDebug *log.Logger
	logInfo  *log.Logger
	logError *log.Logger
	lbJack   *lumberjack.Logger

	logFilename string
}

var _pLog *logger

func newLogger() *logger {
	if _pLog == nil {
		_pLog = &logger{}
	}
	return _pLog
}

func getLogger() *logger {
	return _pLog
}

func SetLogLevel(conf *configure.ValueLog) {
	pLog := newLogger()
	pLog.disableDebug = !conf.EnableDebug
	pLog.disableInfo  = !conf.EnableInfo
	pLog.disableError = !conf.EnableError
}

func GetLogWriter() io.Writer {
	pLog := getLogger()
	return pLog.lbJack
}

func InitLogger(conf *configure.ValueLog) {
	pLog := newLogger()
	var err error

	SetLogLevel(conf)
	pLog.logFilename = conf.LogFile
	dirName := filepath.Dir(conf.LogFile)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
	pLog.fpLog, err = os.OpenFile(pLog.logFilename, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	//defer pLog.fpLog.Close()

	pLog.logDebug = log.New(pLog.fpLog, "[D] ", log.Ldate | log.Lmicroseconds)
	pLog.logInfo  = log.New(pLog.fpLog, "[I] ", log.Ldate | log.Lmicroseconds)
	pLog.logError = log.New(pLog.fpLog, "[E] ", log.Ldate | log.Lmicroseconds)

	pLog.lbJack = &lumberjack.Logger{
		Filename:   pLog.logFilename,
		MaxSize:    conf.MaxSize,	// megabytes
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,	// days
		LocalTime:  conf.LocalTime,
		Compress:   conf.Compress,	// disabled by default
	}
	pLog.logDebug.SetOutput(pLog.lbJack)
	pLog.logInfo.SetOutput(pLog.lbJack)
	pLog.logError.SetOutput(pLog.lbJack)

	//pLog.logInfo.Printf("Logger Initialization completed, File: %s\n", pLog.logFilename)
}

func Close() {
	pLog := getLogger()
	if pLog.fpLog != nil {
		_ = pLog.fpLog.Close()
		pLog.fpLog = nil
	}
}

func LogRotate() {
	pLog := getLogger()
	Logd("Log File Rotate!")
	_ = pLog.lbJack.Rotate()
}

func Logd(format string, v ...interface{})  {
	pLog := getLogger()
	if pLog.logDebug == nil {
		//panic("PANIC: logDebug is nil")
		fmt.Println("ERROR, pLog.logDebug is nil")
		return
	}
	if pLog.disableDebug == true {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	callerInfo := fmt.Sprintf("%s:%d (%s): ", filepath.Base(file), line, runtime.FuncForPC(function).Name())
	pLog.logDebug.Printf(callerInfo + format + "\n", v...)
}

func Logi(format string, v ...interface{})  {
	pLog := getLogger()
	if pLog.logInfo == nil {
		//panic("PANIC: logInfo is nil")
		fmt.Println("ERROR, pLog.logInfo is nil")
		return
	}
	if pLog.disableInfo == true {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	callerInfo := fmt.Sprintf("%s:%d (%s): ", filepath.Base(file), line, runtime.FuncForPC(function).Name())
	s := fmt.Sprintf(callerInfo + format, v...)
	pLog.logInfo.Printf(Yellow(s))
	//pLog.logInfo.Printf(callerInfo + format + "\n", v...)
}

func Loge(format string, v ...interface{})  {
	pLog := getLogger()
	if pLog.logError == nil {
		//panic("PANIC: logError is nil")
		fmt.Println("ERROR, pLog.logError is nil")
		return
	}
	if pLog.disableError == true {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	callerInfo := fmt.Sprintf("%s:%d (%s): ", filepath.Base(file), line, runtime.FuncForPC(function).Name())
	s := fmt.Sprintf(callerInfo + format, v...)
	pLog.logError.Printf(Red(s))
	//pLog.logError.Printf(callerInfo + format + "\n", v...)
}
