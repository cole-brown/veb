package main

import (
	"os"
	"runtime"
	"fmt"
	"time"
	"path/filepath"
)

const (
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
)

// Returns a closure that implements filepath.WalkFn
func walker(out chan string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		// ignore veb metadata folders
		if info.IsDir() && info.Name() == META_FOLDER {
			return filepath.SkipDir
		}

		// spawn off subdirs
//		if info.IsDir() {
//			go filepath.Walk(something)
//		}

		// only files for now.
		// TODO: Possibly also grab symlinks later
		if info.Mode()&os.ModeType != 0 {
			return nil // ignore
		}

		out <- path

		return err
	}
}

func main() {
	fmt.Println("num CPUs:", runtime.NumCPU())

	NUM_GOR := runtime.NumCPU()

	// setup goroutine parallization
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)

	c := make(chan string, NUM_GOR)

	walkroot := "test/data/walk"
//	visit := func(path string, info os.FileInfo, err error) error {
//		c <- path
//		return err
//	}

	go func() {
		filepath.Walk(walkroot, walker(c))
		close(c)
	}()

	t0 := time.Now()
	numf := 0
	for i := range c {
		i = i + " "
		numf++
	}
	t1 := time.Now()

	fmt.Printf("\n\n%v for %v files\n", t1.Sub(t0), numf)
}

////////////////////////////////////////////////////////////////////////////////
func oldmain() {
	fmt.Println("num CPUs:", runtime.NumCPU())

	// setup goroutine parallization
//	runtime.GOMAXPROCS(runtime.NumCPU())

	const (
		NUM = 5000000
	)

	c := make(chan int, NUM)

	for i := 0; i < NUM; i++ {
		go func(x int) {
			c <- 1
//			fmt.Print(".")
		}(i)
	}

	t0 := time.Now()
	sum := 0
	for i := 0; i < NUM; i++ {
		sum += <-c
	}
	t1 := time.Now()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("\n\n%v to do %v things\n", t1.Sub(t0), NUM)


}

