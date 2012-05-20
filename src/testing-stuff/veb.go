// MAIN SYSTEMS ONLINE

package main

import (
	"runtime"
	"os"
//	"path/filepath"
	"log"
	"fmt"
	"crypto"
//	"time"
	"veb"
)

func main() {
	var timer veb.Timer
	timer.Start()

	// setup goroutine parallization
	// cuz it just runs on one processor out of the box...
	// TODO: is running on all procs a good idea? (will it starve the system?)
	NUM_CPUS := runtime.NumCPU()
	runtime.GOMAXPROCS(NUM_CPUS)
	CHAN_SIZE := 1000
	maxHandlers := NUM_CPUS
	if NUM_CPUS > 4 {
		maxHandlers /= 2
	}

	// cd to data's base folder so relative paths in index are nice nice.
	testroot := "test/scratch"
	os.Chdir(testroot)
	root := "."

	// new index
	// ignoring Remote for now
	errlog := log.New(os.Stderr, "veb: ", log.LstdFlags|log.Lshortfile)
	index := veb.New("foo", crypto.SHA1, errlog)

	// initial scan of files
	newfiles := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(root, newfiles)

	// handle new/changed files
	// all new for now, so just xsum & add
	updates := make(chan veb.IndexEntry, CHAN_SIZE)
	quit := make(chan int, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		go xsumHandler(newfiles, updates, quit, errlog)
	}
	done := make(chan int)
	go update(index, updates, done, errlog)

	// wait for handlers to finish before wrapping up
	for i := 0; i < maxHandlers; i++ {
		<-quit
	}
	close(updates)
	<-done

	for _, i := range index.Files {
		fmt.Printf("%x %s\n", i.Xsum, i.Path)
	}

	timer.Stop()
	fmt.Println("\nveb took", timer.Duration())
}

func xsumHandler(files, updates chan veb.IndexEntry, quit chan int, log *log.Logger) {
	for f := range files {
		// calculate checksum hash
		err := veb.Xsum(f, updates, log)
		if err != nil {
			log.Println("Xsum failed:", err)
		}
	}
	quit <- 1
}

func update(index *veb.Index, updates chan veb.IndexEntry, quit chan int, log *log.Logger) {
	for up := range updates {
		err := index.Update(up)
		if err != nil {
			log.Println("Update failed:", err)
		}
	}
	quit <- 1
}


func oldmain() {
//	testfile := "test/data/hash/rand2mb.bin"
//	testfile := "test/data/short-walk"

	// channel for files and the checksum results
//	in  := make(chan FileInfo)
//	out := make(chan []byte)

//	sum, err := veb.Xsum(testfile, nil, nil)
//	if (err == nil) {
//		// shasum format
//		fmt.Printf("%x  %s\n", sum, testfile)
//	} else {
//		fmt.Println("Xsum failed:", err)
//	}

//	err  := filepath.Walk(testfile, veb.Xsum)
//	fmt.Println("\nwalk returned", err)
}

// TODO
//  - statistics!
//    - time, etc
// - unit test!
// - better errors!