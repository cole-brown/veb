package veb

import (
	"log"
)

const (
	// debug help
	IN_FUNC  = "ENTERING"
	OUT_FUNC = "LEAVING "

	// prefixes
	P_ERR = "error >> "
	P_WRN = "warn  >> "
	P_IFO = "info  >> "
)

type Log struct {
	*log.Logger
}

func NewLog(log *log.Logger) *Log {
	return &Log{log}
}

func (l *Log) Err() *Log {
	l.SetPrefix(P_ERR)
	return l
}

func (l *Log) Warn() *Log {
	l.SetPrefix(P_WRN)
	return l
}

func (l *Log) Info() *Log {
	l.SetPrefix(P_IFO)
	return l
}

func (l *Log) Trace(s string) string {
	l.Info().Println(IN_FUNC, s)
	return s
}

func (l *Log) Un(s string) {
	l.Info().Println(OUT_FUNC, s)
}