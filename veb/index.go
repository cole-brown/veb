// Copyright 2012 The veb Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Index maintains an index of files & file info. It can Save() and Load() this
// index to & from the index file in the veb metadata directory.
//
// Main function of interest is Check() which quickly checks the directories for
// new/modified files. Once these files have been processed, Update() needs to be 
// called to update the Index with their new checksums and file info.
//
// Index is not responsible for checksumming files, and does compute xsusm
// to determine if a file's changed. It looks at file stats instead.

package veb

import (
	"crypto"
	"encoding/gob"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
	XSUMS_FILE  = "xsums" // inside of META_FOLDER only
	LOG_FILE    = "log.txt"// inside of META_FOLDER only
)

// The veb index is a map indexed by relative paths
// (paths start at the dir where the .veb directory is located)
type Index struct {
	Files  map[string]IndexEntry
	Remote string      // absolute path to backup location root
	Hash   crypto.Hash // hash function used. 0 = not yet hashed
	Root   string      // root of this veb repository
	log    *Log        // error/warn/info logging
}

// A veb index entry/value
// TODO: rename to just Entry
type IndexEntry struct {
	Path string // filepath, same as the entry's key in the Index map
	Xsum []byte // checksum of file
	// Stat info
	// can't just hang onto os.FileInfo, because it's actually an os.fileStat
	// and that has no exported fields
	Name    string      // base name of file
	Size    int64       // length in bytes
	Mode    os.FileMode // file mode bits
	ModTime time.Time   // modification time
}

// Creates a new, empty, Index
func New(hash crypto.Hash, root string) *Index {
	ret := Index{make(map[string]IndexEntry), "", hash, root, nil}
	return &ret
}

// Reads the index in from the index file, decodes with gob into new Index
func Load(root string, log *Log) (*Index, error) {
	// open index file
	file, err := os.Open(path.Join(root, META_FOLDER, INDEX_FILE))
	if err != nil {
		log.Err().Println(err)
		return nil, err
	}
	defer file.Close() // make sure to close that file

	// Decode the index
	var ret Index
	dec := gob.NewDecoder(file)
	err = dec.Decode(&ret)
	if err != nil {
		log.Err().Println("couldn't load index:", err)
		return nil, err
	}

	// Attach the logger and root
	ret.log = log
	ret.Root = root

	return &ret, nil
}

// Saves index to file, encoded with gob
func (x *Index) Save() error {
	// move previous index file to backup file, in case something goes badly.
	// overwrites previous backup index, if it exists.
	err := os.Rename(path.Join(x.Root, META_FOLDER, INDEX_FILE),
		path.Join(x.Root, META_FOLDER, INDEX_FILE+"~"))
	if err != nil {
		x.log.Warn().Println("could not backup old index:", err)
		// Don't return error. Only errors if index doesn't exist, which is fine
		// becaus we're about to save a new one.
	}

	// open index file
	// overwrites if already exists
	file, err := os.Create(path.Join(x.Root, META_FOLDER, INDEX_FILE))
	if err != nil {
		x.log.Err().Println(err)
		return err
	}
	defer file.Close() // make sure to close that file

	// send index to file
	enc := gob.NewEncoder(file)
	err = enc.Encode(x)
	if err != nil {
		x.log.Err().Println("couldn't save index:", err)
		return err
	}

	// move previous index file to backup file, in case something goes badly.
	// overwrites previous backup index, if it exists.
	err = os.Rename(path.Join(x.Root, META_FOLDER, XSUMS_FILE),
		path.Join(x.Root, META_FOLDER, XSUMS_FILE+"~"))
	if err != nil {
		x.log.Warn().Println("could not backup old xsums:", err)
		// Don't return error. 
	}

	// new xsums file
	xsfile, err := os.Create(path.Join(x.Root, META_FOLDER, XSUMS_FILE))
	if err != nil {
		x.log.Err().Println(err)
		return err
	}
	defer xsfile.Close() // make sure to close that file
	
	// write all xsums out
	for _, e := range x.Files {
		xsfile.WriteString(XsumString(&e))
	}

	return nil
}

// Checks file stats against stats in the index; does not recompute checksum.
// If file differs, returns false. 
// If file does not exist in Index, returns false with an IndexError.
func (x Index) Check(changed chan IndexEntry) error {
//	// save prev dir
//	prevDir, err := os.Getwd()
//	if err != nil {
//		x.log.Err().Println(err)
//		return err
//	}
//
//	// cd to root so paths are relative
//	err = os.Chdir(root)
//	if err != nil {
//		x.log.Err().Println(err)
//		return err
//	}

	// find changes
	err := filepath.Walk(x.Root, x.checkWalker(changed))
	if err != nil {
		x.log.Err().Println(err)
	}
	close(changed)

//	// return to prev dir
//	err = os.Chdir(prevDir)
//	if err != nil {
//		x.log.Err().Println(err)
//	}
	return err
}

// Get file stats and save to entry
func SetStats(root string, entry *IndexEntry) error {
	// get file's size & such
	info, err := os.Lstat(path.Join(root, entry.Path))
	if err != nil {
		return err
	}

	// update entry fields (xsum, path already done)
	entry.Name    = info.Name()
	entry.Size    = info.Size()
	entry.Mode    = info.Mode()
	entry.ModTime = info.ModTime()

	return nil
}

// File has been delt with; update xsum and file stats in Index
func (x Index) Update(entry *IndexEntry) error {
	err := SetStats(x.Root, entry)
	if err != nil {
		x.log.Err().Println(err)
		return err
	}

	// Add/Update entry in Index
	x.Files[entry.Path] = *entry

	return nil
}

// Returns a closure that implements filepath.WalkFn
// checkWalker's closure checks files encountered against those in the index
func (x Index) checkWalker(changed chan IndexEntry) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// ignoring errors so we can continue if possible
			x.log.Err().Println(err)
			return nil
		}

		// ignore veb metadata folders
		if info.IsDir() && info.Name() == META_FOLDER {
			return filepath.SkipDir
		}

		// only files for now.
		// TODO: Possibly also grab symlinks later
		if info.Mode()&os.ModeType != 0 {
			return nil // ignore
		}

		// make path relative
		path = strings.Replace(path, x.Root+"/", "", 1)

		// compare current file stats against index's stats
		file, ok := x.Files[path]
		if !ok {
			// not in index (new file)
			// add to index (w/ no stats)
			x.Files[path] = IndexEntry{Path: path}

			// add to channel for processing
			changed <- x.Files[path]
		} else if file.Size != info.Size() ||
			file.ModTime != info.ModTime() ||
			file.Mode != info.Mode() {
			// modified file
			// add to channel for processing
			changed <- x.Files[path]
		}

		return err
	}
}

// TODO
//  - notice deleted files!
//  - statistics!
//    - time, etc
//  - non-main!
//  - unit test!
