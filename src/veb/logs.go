package veb

import (
	"log"
	"os"
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