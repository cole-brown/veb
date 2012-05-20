// hashing files!
 
package veb

import (
	"io"
	"os"
	"log"
	"crypto/sha1"
	// TODO: these hashes
	//	"crypto/sha256"
	//	"crypto/md5"
//	"fmt"
)

func Xsum(entry IndexEntry, updates chan IndexEntry, log *log.Logger) error {
	// TODO: make hasher from supplied crypto.Hash
	// - crypto.Available(), crypto.New()
	hasher := sha1.New()

	file, err := os.Open(entry.Path)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(hasher, file)
	if err != nil {
		log.Println(err)
		return err
	}
	// TODO: check file size to make sure all read?

	entry.Xsum = hasher.Sum(nil)
	updates <- entry

//	fmt.Printf("%x  %s\n", hasher.Sum(nil), entry.Path)
	
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