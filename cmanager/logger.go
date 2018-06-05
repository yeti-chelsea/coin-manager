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
	LogLevels []LOG_LEVEL
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

	lgr.LogLevels = make([]LOG_LEVEL, 3)
	return true
}

func (lgr *Logger) SetLogLevel(log_chl int, level LOG_LEVEL) {
	lgr.LogLevels[log_chl] = level
}

func (lgr *Logger) Debug(log_chl int, v ...interface{}) {
	if lgr.LogLevels[log_chl] >= DEBUG_LEVEL {
		lgr.DEBUG.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Info(log_chl int, v ...interface{}) {

	if lgr.LogLevels[log_chl] >= INFO_LEVEL {
		lgr.INFO.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Warning(log_chl int, v ...interface{}) {

	if lgr.LogLevels[log_chl] >= WARNING_LEVEL {
		lgr.WARNING.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) Error(log_chl int, v ...interface{}) {

	if lgr.LogLevels[log_chl] >= ERROR_LEVEL {
		lgr.ERROR.Output(2, fmt.Sprint(v...))
	}
}

func (lgr *Logger) ChangeLogLevel(log_chl int) {
	if lgr.LogLevels[log_chl] != DEBUG_LEVEL {
		lgr.LogLevels[log_chl] = DEBUG_LEVEL
	}else {
		lgr.LogLevels[log_chl] = INFO_LEVEL
	}
}
