// veb commands
//
// valid commands are:
//   init

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"veb"
//	"runtime"
//	"path/filepath"
//	"crypto"
//	"time"
)

const (
	INIT = "init"

	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
	XSUMS_FILE  = "xsums" // inside of META_FOLDER only
)

func main() {
	out := log.New(os.Stdout, "", 0)
	errlog := log.New(os.Stderr, "veb: ", log.LstdFlags|log.Lshortfile)

	// define flags
	// TODO

	// parse flags & args
	flag.Parse()

	// sanity check
	if len(flag.Args()) == 0 {
		// TODO
		out.Fatal("INSERT HELP HERE")
	}

	// act on command
	switch flag.Args()[0] {
	case INIT:
		err := Init(errlog)
		if err != nil {
			out.Fatal(err)
		}
		pwd, _ := os.Getwd()
		out.Println("Initialized empty veb repository at", pwd)

	default:
		// TODO
		fmt.Println("INSERT HELP HERE")
	}
}

// creates veb's META_FOLDER in current directory and empty veb metadata 
// files inside META_FOLDER
func Init(log *log.Logger) error {
	var timer veb.Timer
	timer.Start()

	// create veb dir
	err := os.Mkdir(META_FOLDER, 0755)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("veb could not create metadata directory: %v", err)
	}

	// create index file
	index, err := os.Create(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		log.Println(err)
		return fmt.Errorf("veb could not create metadata index file: %v", err)
	}
	defer index.Close()

	// create xsums file
	xsums, err := os.Create(path.Join(META_FOLDER, XSUMS_FILE))
	if err != nil {
		log.Println(err)
		return fmt.Errorf("veb could not create metadata xsums file: %v", err)
	}
	defer xsums.Close()

	// create & save empty index
	// TODO ignoring remote for now
//	index := veb.New("foo", crypto.SHA1, errlog)
//	err = index.Save()
//	if err != nil {
//		return err // logged in save
//	}
//
	timer.Stop()
	log.Println("init took", timer.Duration())
	return nil
}
