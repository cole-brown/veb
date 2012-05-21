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

func Xsum(entry *IndexEntry, logs *Logs) error {
	// TODO: make hasher from supplied crypto.Hash
	// - crypto.Available(), crypto.New()
	hasher := sha1.New()

	file, err := os.Open(entry.Path)
	if err != nil {
		logs.Err.Println(err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(hasher, file)
	if err != nil {
		logs.Err.Println(err)
		return err
	}
	// TODO: check file size to make sure all read?

	entry.Xsum = hasher.Sum(nil)

	return err
}

func XsumString(entry *IndexEntry) string {
	// shasum format:
	return fmt.Sprintf("%x  %s\n", entry.Xsum, entry.Path)
}

// TODO
//  - hashing options!
//    - make a classy thingy 
//  - statistics!
//    - time, etc
// - non-main!
// - unit test!
// - better errors!