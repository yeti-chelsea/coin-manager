package cmanager

import (
    "log"
	"fmt"
	"io"
)

type Logger struct {
    DEBUG   *log.Logger
    INFO    *log.Logger
    WARNING *log.Logger
    ERROR   *log.Logger
}

func (lgr *Logger) InitLogger(file_ptr *io.Writer) bool {

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

	return true
}

func (lgr *Logger) Debug(v ...interface{}) {
	lgr.DEBUG.Output(2, fmt.Sprint(v...))
}

func (lgr *Logger) Info(v ...interface{}) {
	lgr.INFO.Output(2, fmt.Sprint(v...))
}

func (lgr *Logger) Warning(v ...interface{}) {
	lgr.WARNING.Output(2, fmt.Sprint(v...))
}

func (lgr *Logger) Error(v ...interface{}) {
	lgr.ERROR.Output(2, fmt.Sprint(v...))
}
