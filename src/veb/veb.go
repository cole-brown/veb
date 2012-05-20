// veb commands
//
// valid commands are:
//   init

package main

import (
	"crypto"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"
	"veb"

//	"runtime"
//	"path/filepath"
//	"crypto"
//	"time"
)

const (
	// commands
	FIX    = "fix"
	HELP   = "help"
	PUSH   = "push"
	PULL   = "pull"
	SYNC   = "sync"
	INIT   = "init"
	STATUS = "status"
	VERIFY = "verify"
	REMOTE = "remote"

	// debug help
	// TODO: move to Logs?
	IN_FUNC  = "START"
	OUT_FUNC = "END"

	// TODO: move somewhere else. Index uses thes too.
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
	XSUMS_FILE  = "xsums" // inside of META_FOLDER only
)

// pretty print filesizes
// shamelessly stolen from Effective Go: http://golang.org/doc/effective_go.html#constants
type ByteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

// Command veb to do things with your stuff!
// See all the commands in const, above. Also flags.
// 'veb help' for a pretty print of flags/commands on command line.
func main() {
	out := log.New(os.Stdout, "", 0)
	logs := veb.NewLogs()

	// define flags
	// TODO

	// parse flags & args
	flag.Parse()

	// sanity check
	if len(flag.Args()) == 0 {
		// TODO
		out.Fatal("INSERT HELP HERE")
	}

	// act on command
	switch flag.Args()[0] {
	case INIT:
		err := Init(logs)
		if err != nil {
			out.Fatal(err)
		}
		pwd, _ := os.Getwd()
		out.Println("Initialized empty veb repository at", pwd)

	case STATUS:
		err := Status(logs)
		if err != nil {
			out.Fatal(err)
		}

	case FIX:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case PUSH:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case PULL:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case SYNC:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case VERIFY:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case REMOTE:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case HELP:
		// may provide per-cmd help later, but for now, just the default help
		// TODO: per command help
		fallthrough
	default:
		// TODO
		fmt.Println("INSERT HELP HERE")
	}
}

// creates veb's META_FOLDER in current directory and empty veb metadata 
// files inside META_FOLDER
func Init(logs *veb.Logs) error {
	logs.Info.Println(IN_FUNC, INIT)
	var timer veb.Timer
	timer.Start()

	// create veb dir
	err := os.Mkdir(META_FOLDER, 0755)
	if err != nil {
		logs.Err.Println(err)
		return fmt.Errorf("veb could not create metadata directory: %v", err)
	}

	// create index file
	indexf, err := os.Create(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		logs.Err.Println(err)
		return fmt.Errorf("veb could not create metadata index file: %v", err)
	}
	indexf.Close()

	// create xsums file
	xsums, err := os.Create(path.Join(META_FOLDER, XSUMS_FILE))
	if err != nil {
		logs.Err.Println(err)
		return fmt.Errorf("veb could not create metadata xsums file: %v", err)
	}
	xsums.Close()

	// create & save empty index
	// TODO ignoring remote for now
	index := veb.New("foo", crypto.SHA1, logs)
	err = index.Save()
	if err != nil {
		return err // logged in save
	}

	timer.Stop()
	logs.Info.Println(OUT_FUNC, INIT)
	logs.Info.Println(INIT, "took", timer.Duration())
	return nil
}

// Check for updated/new files in repo, then nicely print out results in 
// the following format:
//
//    veb repository at /path/to/here
//  
//    --------------
//    Changed files:
//    --------------
//      foo/bar/baz.bin
//          - filesize increased 400 bytes (90.3MB -> 90.3MB)
//          - modified on (2012-05-19 16:11:05)
//  
//      foo/quux.mp3  
//          - modification time only (2012-05-19 16:11:05)
//  
//      
//    ----------
//    New files:
//    ----------
//      xyzzy.iso  
//          - 8.9GB, modified on (2012-05-19 16:11:05)
//  
//      firefly.m4v
//          - 80MB, modified on (2012-05-19 16:11:05)
//  
//    MAKE SURE CHANGED FILES ARE THINGS YOU'VE ACTUALLY CHANGED
//      (use "veb fix <file>" if a file has been corrupted in this repository)
//      (use "veb push", "veb pull", or "veb sync" to commit changed/new files)
//
// Format may seem a litte overly whitespaced, but when large amounts of files
// are present (and/or long path/file names), the space is nice.
func Status(logs *veb.Logs) error {
	logs.Info.Println(IN_FUNC, STATUS)
	var timer veb.Timer
	timer.Start()

	// TODO: move this to own function cdBase(), or something
	// search down the present working dir for META_FOLDER && cd there
	MAX_PARENTS := 15
	pwd, err := os.Getwd()
	if err != nil {
		logs.Err.Println(err)
		return fmt.Errorf("veb could not figure out where it is: %v", err)
	}
	dir := pwd
	found := false
	for i := 0; i < MAX_PARENTS; i++ {
		// check current
		fi, err := os.Stat(path.Join(dir, META_FOLDER))
		if err != nil {
			// If the dir doesn't exist, fine.
			// Log other errors.
			if !os.IsNotExist(err) {
				logs.Err.Println(err)
				return fmt.Errorf("veb could not find %v metafolder: %v", META_FOLDER, err)
			}
		} else if fi.IsDir() {
			// dir is now at base directory
			found = true
			os.Chdir(dir)
			break
		}

		// if not found, check parent dir next time
		dir = path.Dir(dir)
	}
	if !found {
		logs.Err.Println("veb could not find", META_FOLDER, "metafolder")
		return fmt.Errorf("veb could not find %v metafolder at or %v folders below: %v",
			META_FOLDER, MAX_PARENTS, pwd)
	}

	// load the index
	index, err := veb.Load(logs)
	if err != nil {
		return fmt.Errorf("veb could not load index")
	}

	// setup goroutine parallization
	// cuz it just runs on one processor out of the box...
	// TODO: is running on all procs a good idea? (will it starve the system?)
	// TODO: move elsewhere
	NUM_CPUS := runtime.NumCPU()
	runtime.GOMAXPROCS(NUM_CPUS)
	CHAN_SIZE := 1000
	maxHandlers := NUM_CPUS
	if NUM_CPUS > 4 {
		maxHandlers /= 2
	}

	// check for changes
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(".", files)

	// parse into new vs changed
	newFiles := make([]string, 0)
	changedFiles := make([]string, 0)
	var noTime time.Time
	for f := range files {
		// new file?
		if f.Name == "" && f.Size == 0 && f.Mode == 0 && f.ModTime == noTime {
			// new file.
			newFiles = append(newFiles, f.Path)
		} else {
			// changed file.
			changedFiles = append(changedFiles, f.Path)
		}
	}

	// print intro
	fmt.Println("veb repository at", dir, "\n")

	INDENT_F := " " // use with Println == 2 spaces
	INDENT_I := "      -"

	// print new files
	if len(newFiles) > 0 {
		fmt.Println("----------")
		fmt.Println("New files:")
		fmt.Println("----------")
		for _, f := range newFiles {
			// print file name
			fmt.Println(INDENT_F, f)

			// get stats for file info line
			fi, err := os.Stat(f)
			if err != nil {
				logs.Err.Println(err)
				// print file info oops
				fmt.Println(INDENT_I, "could not get file info")
			} else {
				// print file info
				size := ByteSize(fi.Size())
				fmt.Printf("%s %s, modified on (%v)\n", INDENT_I, size, fi.ModTime())
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	// print changed files
	if len(changedFiles) > 0 {
		fmt.Println("--------------")
		fmt.Println("Changed files:")
		fmt.Println("--------------")
		for _, f := range changedFiles {
			// print file name
			fmt.Println(INDENT_F, f)

			// get stats for file info line
			fi, err := os.Stat(f)
			if err != nil {
				logs.Err.Println(err)
				// print file info oops
				fmt.Println(INDENT_I, "could not get file info")
			} else {
				// figure out filesize
				curSize := ByteSize(fi.Size())
				prevSize := ByteSize(index.Files[f].Size)
				sizeChange := curSize - prevSize
				direction := "increased"
				if sizeChange < 0 {
					direction = "decreased"
					sizeChange -= sizeChange // absolute value
				}

				sanity := false

				// print size change
				if sizeChange != 0 {
					fmt.Printf("%s filesize %s %s (%s -> %s)",
						INDENT_I, direction, sizeChange, curSize, prevSize)
					sanity = true
				}

				// print mtime
				if index.Files[f].ModTime != fi.ModTime() {
					fmt.Printf("%s modified on (%v)\n", INDENT_I, fi.ModTime())
					sanity = true
				}

				// print mode
				if index.Files[f].Mode != fi.Mode() {
					fmt.Printf("%s file mode changed (%3o-> %3o)\n",
						INDENT_I, index.Files[f].Mode, fi.Mode())
					sanity = true
				}

				// sanity check & snark
				if !sanity {
					fmt.Printf("%s ...well /something/ changed. Dunno what. *shrugs*\n", INDENT_I)
				}
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	// print outro
	if len(changedFiles) == 0 && len(newFiles) == 0 {
		fmt.Println("No changes or new files.")
	} else {
		fmt.Println("MAKE SURE CHANGED FILES ARE THINGS YOU'VE ACTUALLY CHANGED")
		fmt.Println("  (use \"veb fix <file>\" if a file has been corrupted in this repository)")
		fmt.Println("  (use \"veb push\", \"veb pull\", or \"veb sync\" to commit changed/new files)")
	}
	fmt.Printf("\nsummary: %d new, %d changed\n", len(newFiles), len(changedFiles))

	timer.Stop()
	logs.Info.Println(OUT_FUNC, STATUS)
	logs.Info.Printf("%s (%d new, %d changed) took %v\n",
		STATUS, len(newFiles), len(changedFiles), timer.Duration())
	return nil
}
