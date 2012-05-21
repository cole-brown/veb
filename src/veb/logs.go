package veb

import (
	"log"
	"os"
)

const (
	// debug help
	IN_FUNC  = "ENTERING"
	OUT_FUNC = "LEAVING "
)

type Logs struct {
	Err  *log.Logger // error-level logging
	Warn *log.Logger // warning-level logging
	Info *log.Logger // info-level logging
}

func NewLogs() *Logs {
	ret := Logs{
		log.New(os.Stderr, "error >> ", log.LstdFlags|log.Lshortfile),
		log.New(os.Stderr, "warn  >> ", log.LstdFlags|log.Lshortfile),
		log.New(os.Stderr, "info  >> ", log.LstdFlags|log.Lshortfile)}
	return &ret
}

func (l *Logs) Trace(s string) string {
	l.Info.Println(IN_FUNC, s)
	return s
}

func (l *Logs) Un(s string) {
	l.Info.Println(OUT_FUNC, s)
}