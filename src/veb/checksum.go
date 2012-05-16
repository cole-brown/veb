// hashing files!
 
package veb

import (
	"io"
	"os"
	"crypto/sha1"
	// TODO: these hashes
	//	"crypto/sha256"
	//	"crypto/md5"
	"fmt"
)

func Xsum(path string, info os.FileInfo, err error) error {// ([]byte, error) {
	// only files for now.
	// TODO: Possibly also grab symlinks later
	if info.Mode() & os.ModeType != 0 {
		return nil // ignore
	}

	hasher := sha1.New()

	file, err := os.Open(path)
	if err == nil {
		_, err = io.Copy(hasher, file)
		// TODO: check file size to make sure all read?
		fmt.Printf("%x  %s\n", hasher.Sum(nil), path)
	}
	
//	return hasher.Sum(nil), err
	return err
}

// shasum format:
//		fmt.Printf("%x  %s\n", hasher.Sum(nil), testfile)

// TODO
//  - hashing options!
//    - make a classy thingy 
//  - statistics!
//    - time, etc
// - non-main!
// - unit test!
// - better errors!