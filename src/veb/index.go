// TODO: package doc
// Meta-file stuff!
// Index.go does not compute file checksums.

// TODO: package veb/index
package main

import (
	"crypto"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
)

// The veb index is a map indexed by relative paths
// (paths start at the dir where the .veb directory is located)
type Index struct {
	Files  map[string]IndexEntry
	Remote string      // absolute path to backup location root
	Hash   crypto.Hash // hash function used. 0 = not yet hashed
	log    *log.Logger // error logging
}

// A veb index entry/value
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

func main() {
	//	testfile := "test/data/hash/rand2mb.bin"
	//	testindex := "test/data/hash/index.veb"
	//
	//	// stat data file
	//	info, err := os.Lstat(testfile)
	//	if err == nil {
	//		fmt.Printf("%+v\n", info)
	//	}
	//
	//	// open index file
	//	index, err := os.Create(testindex) // overwrites if already exists
	//	if err != nil {
	//		fmt.Println("\nError'd!", err)
	//		return
	//	}
	//	defer index.Close()
	//
	//	// write to index
	//	_, err = index.WriteString("Hello, World. - " + testfile)
	//	if err != nil {
	//		fmt.Println("\nError'd!", err)
	//		return
	//	}
	//
	//	if err != nil {
	//		fmt.Println("\nError'd!", err)
	//	}

//	vi := make(map[string]IndexEntry)
//	vr := "/foo/bar"
//	vl := log.New(os.Stderr, "veb: ", log.LstdFlags|log.Lshortfile)
//	veb := Index{vi, vr, 0, vl}
//
//	fmt.Println(os.Getwd())
//	root := "test/data/index-simple"
//	os.Chdir(root)
//	root = "."
//
//	err := veb.build(root)
//	if err != nil {
//		fmt.Println("\nError'd!", err)
//	}
//	fmt.Println(len(veb.Files))
//	//	fmt.Println(veb)
//
//	c := make(chan string, 10)
//	err = veb.Check(root, c)
//	if err != nil {
//		fmt.Println("\nError'd!", err)
//	}
//	for i := 0; i < len(c); i++ {
//		fmt.Println(<-c)
//	}
//
//	err = veb.Save()
//	if err != nil {
//		fmt.Println("\nError'd!", err)
//	}
//
//	veb2, err := Load(vl)
//	if err != nil {
//		fmt.Println("\nError'd!", err)
//	}
//	fmt.Println(veb)
//	fmt.Println(veb2)
}

// Creates a new, empty, Index
func New(root, remote string, hash crypto.Hash, log *log.Logger) *Index {
	ret := Index{make(map[string]IndexEntry), remote, hash, log}
	return ret
}

// Reads the index in from the index file, decodes with gob into new Index
func Load(log *log.Logger) (*Index, error) {
	// open index file
	file, err := os.Open(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer file.Close() // make sure to close that file

	// Decode the index
	var ret Index
	dec := gob.NewDecoder(file)
	err = dec.Decode(&ret)
	if err != nil {
		log.Println("couldn't load index:", err)
		return nil, err
	}

	// Attach the logger
	ret.log = log

	return &ret, nil
}

// Saves index to file, encoded with gob
func (x *Index) Save() error {
	// move previous index file to backup file, in case something goes badly.
	// overwrites previous backup index, if it exists.
	err := os.Rename(path.Join(META_FOLDER, INDEX_FILE),
		path.Join(META_FOLDER, INDEX_FILE+"~"))
	if err != nil {
		x.log.Println("could not backup old index:", err)
		// Don't return error. Only errors if index doesn't exist, which is fine
		// becaus we're about to save a new one.
	}

	// open index file
	// overwrites if already exists
	file, err := os.Create(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		x.log.Println(err)
		return err
	}
	defer file.Close() // make sure to close that file

	// send index to file
	enc := gob.NewEncoder(file)
	err = enc.Encode(x)
	if err != nil {
		x.log.Println("couldn't save index:", err)
		return err
	}

	return nil
}

// Checks file stats against stats in the index; does not recompute checksum.
// If file differs, returns false. 
// If file does not exist in Index, returns false with an IndexError.
func (x Index) Check(root string, changed chan string) error {
	err := filepath.Walk(root, x.checkWalker(changed))

	if err != nil {
		x.log.Println(err)
	}
	return err
}

// File has been delt with; update xsum and file stats in Index
func (x Index) Update(filepath string, xsum []bytes) error {
	// get file's size & such
	info, err := os.Lstat(filepath)
	if err != nil {
		x.log.Println(err)
		return err
	}

	// Add/Update entry in Index
	x.Files[path] = IndexEntry{
		filepath,
		xsum,
		info.Name(),
		info.Size(),
		info.Mode(),
		info.ModTime()}	
}

// Returns a closure that implements filepath.WalkFn
// checkWalker's closure checks files encountered against those in the index
func (x Index) checkWalker(changed chan string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// ignoring errors so we can continue if possible
			x.log.Println(err)
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

		// compare current file stats against index's stats
		// TODO: also compare mode?
		if file, ok := x.Files[path]; !ok ||
			file.Size != info.Size() ||
			file.ModTime != info.ModTime() {
			// file differs or is not in index
			// add to channel for processing
			changed <- path
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
