// MAIN SYSTEMS ONLINE

package main

import (
	"fmt"
	"path/filepath"
	"veb"
)

func main() {
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