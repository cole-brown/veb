// Copyright 2012 The veb Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// checksums the given entry

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

// checksum of supplied entry is added to the entry itself
// TODO: take in root string
func Xsum(entry *IndexEntry, log *Log) error {
	// TODO: make hasher from supplied crypto.Hash
	// - crypto.Available(), crypto.New()
	hasher := sha1.New()

	file, err := os.Open(entry.Path)
	if err != nil {
		log.Err().Println(err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(hasher, file)
	if err != nil {
		log.Err().Println(err)
		return err
	}

	entry.Xsum = hasher.Sum(nil)

	return err
}

// returns xsum in shasum formatted string (<ASCII hex hash> <filepath>)
func XsumString(entry *IndexEntry) string {
	// shasum format
	return fmt.Sprintf("%x  %s\n", entry.Xsum, entry.Path)
}

// TODO
//  - hashing options!
//  - unit test!
