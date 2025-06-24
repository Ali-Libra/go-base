package logger

import "fmt"

var log LogInterface

const (
	LOGTYPE_FILE    = "file"
	LOGTYPE_CONSOLE = "console"
)

/*
file, "初始化一个文件日志实例"
console, "初始化console日志实例"
*/
// func InitLogger(config map[string]string) (err error) {
// 	log, err = NewFileLogger(config)
// 	if err != nil {
// 		return err
// 	}
// 	log.Init()

// 	return
// }

func CloseLogger() {
	if log != nil {
		log.Close()
	}
}

func InitLogger(name string, config map[string]string) (err error) {
	switch name {
	case LOGTYPE_FILE:
		log, err = NewFileLogger(config)
		if err != nil {
			return err
		}
		log.Init()
	case LOGTYPE_CONSOLE:
		log, err = NewConsoleLogger(config)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf("unsupport logger name:%s", name)
	}

	return
}

func Debug(format string, args ...interface{}) {
	log.Debug(format, args...)
}

func Trace(format string, args ...interface{}) {
	log.Trace(format, args...)
}

func Info(format string, args ...interface{}) {
	log.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	log.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	log.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	log.Fatal(format, args...)
}
