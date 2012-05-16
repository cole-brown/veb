// printing out filenames!

package main

import (
	"path/filepath"
	"os"
	"fmt"
)

func visit(path string, info os.FileInfo, err error) error {
	fmt.Println("Visited:", path)
  return err
}

func Walk(root string, fileFn WalkFunc) {
	// http://golang.org/pkg/path/filepath/#Walk
	err  := filepath.Walk(root, visit)
	fmt.Println("\nwalk returned", err)
}

// TODO
//  - IGNORE DIRECTOIRES!
//  - statistics!
//    - num files walked, time, etc
// - non-main!
// - unit test!
// - better errors!