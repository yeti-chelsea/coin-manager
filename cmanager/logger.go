package cmanager

import (
    "log"
	"fmt"
	"io"
)

type LOG_LEVEL int

const (
	ERROR_LEVEL LOG_LEVEL =  10
	WARNING_LEVEL = 20
	INFO_LEVEL = 30
	DEBUG_LEVEL = 40
)

type Logger struct {
    DEBUG   *log.Logger
    INFO    *log.Logger
    WARNING *log.Logger
    ERROR   *log.Logger

	LogLevel LOG_LEVEL
}

func (lgr *Logger) InitLogger(file_ptr *io.Writer, level LOG_LEVEL) bool {

    lgr.DEBUG = log.New(*file_ptr,
		"DEBUG: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    lgr.INFO = log.New(*file_ptr,
        "INFO: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    lgr.WARNING = log.New(*file_ptr,
        "WARNING: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    lgr.ERROR = log.New(*file_ptr,
        "ERROR: ",
        log.Ldate|log.Ltime|log.Lshortfile)

	lgr.LogLevel = level

	return true
}

func (lgr *Logger) Debug(v ...interface{}) {
	if lgr.LogLevel >= DEBUG_LEVEL {
		lgr.DEBUG.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Info(v ...interface{}) {

	if lgr.LogLevel >= INFO_LEVEL {
		lgr.INFO.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Warning(v ...interface{}) {

	if lgr.LogLevel >= WARNING_LEVEL {
		lgr.WARNING.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Error(v ...interface{}) {

	if lgr.LogLevel >= ERROR_LEVEL {
		lgr.ERROR.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) SetLogLevel() {
	if lgr.LogLevel != DEBUG_LEVEL {
		lgr.LogLevel = INFO_LEVEL
	}else {
		lgr.LogLevel = DEBUG_LEVEL
	}
}
