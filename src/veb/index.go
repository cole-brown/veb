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
	logs   *Logs       // error/warn/info logging
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

// TODO remove
func oldmain() {
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
func New(remote string, hash crypto.Hash, logs *Logs) *Index {
	ret := Index{make(map[string]IndexEntry), remote, hash, logs}
	return &ret
}

// Reads the index in from the index file, decodes with gob into new Index
func Load(logs *Logs) (*Index, error) {
	// open index file
	file, err := os.Open(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		logs.Err.Println(err)
		return nil, err
	}
	defer file.Close() // make sure to close that file

	// Decode the index
	var ret Index
	dec := gob.NewDecoder(file)
	err = dec.Decode(&ret)
	if err != nil {
		logs.Err.Println("couldn't load index:", err)
		return nil, err
	}

	// Attach the logger
	ret.logs = logs

	return &ret, nil
}

// Saves index to file, encoded with gob
func (x *Index) Save() error {
	// move previous index file to backup file, in case something goes badly.
	// overwrites previous backup index, if it exists.
	err := os.Rename(path.Join(META_FOLDER, INDEX_FILE),
		path.Join(META_FOLDER, INDEX_FILE+"~"))
	if err != nil {
		x.logs.Warn.Println("could not backup old index:", err)
		// Don't return error. Only errors if index doesn't exist, which is fine
		// becaus we're about to save a new one.
	}

	// open index file
	// overwrites if already exists
	file, err := os.Create(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		x.logs.Err.Println(err)
		return err
	}
	defer file.Close() // make sure to close that file

	// send index to file
	enc := gob.NewEncoder(file)
	err = enc.Encode(x)
	if err != nil {
		x.logs.Err.Println("couldn't save index:", err)
		return err
	}

	return nil
}

// Checks file stats against stats in the index; does not recompute checksum.
// If file differs, returns false. 
// If file does not exist in Index, returns false with an IndexError.
func (x Index) Check(root string, changed chan IndexEntry) error {
	err := filepath.Walk(root, x.checkWalker(changed))

	if err != nil {
		x.logs.Err.Println(err)
	}
	close(changed)
	return err
}

// File has been delt with; update xsum and file stats in Index
func (x Index) Update(entry IndexEntry) error {
	// get file's size & such
	info, err := os.Lstat(entry.Path)
	if err != nil {
		x.logs.Err.Println(err)
		return err
	}

	// update entry fields (xsum, path already done)
	entry.Name    = info.Name()
	entry.Size    = info.Size()
	entry.Mode    = info.Mode()
	entry.ModTime = info.ModTime()

	// Add/Update entry in Index
	x.Files[entry.Path] = entry

	return nil
}

// Returns a closure that implements filepath.WalkFn
// checkWalker's closure checks files encountered against those in the index
func (x Index) checkWalker(changed chan IndexEntry) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// ignoring errors so we can continue if possible
			x.logs.Err.Println(err)
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
			// add to index (no stats, so if cancelled early, it will show up as 
			// new/modified next time
			x.Files[path] = IndexEntry{Path: path}
			
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
