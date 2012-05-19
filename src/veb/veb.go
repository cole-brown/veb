// MAIN SYSTEMS ONLINE

package main

import (
	"runtime"
	"os"
	"path/filepath"
	"log"
	"fmt"
	"crypto"
//	"time"
	"veb"
)


func main() {
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
	newfiles := make(chan string, CHAN_SIZE)
	go index.Check(root, newfiles)

	// handle new/changed files
	// all new for now, so just xsum & add
	quit := make(chan int, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		go handle(newfiles, quit)
	}
	for i := 0; i < maxHandlers; i++ {
		<-quit // wait for handlers to finish before quitting
	}
}

func handle(files chan string, quit chan int) {
	for f := range files {
		// calculate checksum hash
		err := veb.Xsum(f, nil, nil)
		if err != nil {
			fmt.Println("Xsum failed:", err)
		}
//		if (err == nil) {
//			// shasum format
//			fmt.Printf("%x  %s\n", sum, f)
//		} else {
//			fmt.Println("Xsum failed:", err)
//		}
	}
	quit <- 1
}

func oldmain() {
//	testfile := "test/data/hash/rand2mb.bin"
	testfile := "test/data/short-walk"

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

	err  := filepath.Walk(testfile, veb.Xsum)
	fmt.Println("\nwalk returned", err)
}

// TODO
//  - statistics!
//    - time, etc
// - unit test!
// - better errors!