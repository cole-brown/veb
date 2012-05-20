// veb commands
//
// valid commands are:
//   init

package main

import (
	"bytes"
	"crypto"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"
	"veb"
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

	// misc
	QUIT_RUNE = 'q'
	CHAN_SIZE = 1000

	// debug help
	// TODO: move to Logs?
	IN_FUNC  = "START"
	OUT_FUNC = "END"

	// TODO: move somewhere else. Index uses thes too.
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
	XSUMS_FILE  = "xsums" // inside of META_FOLDER only
)

var (
	maxHandlers int
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
	//  - max CPUs
	//  - verbose

	// parse flags & args
	flag.Parse()

	// setup goroutine parallization
	// cuz it just runs on one processor out of the box...
	// TODO: is running on all procs a good idea? (will it starve the system?)
	NUM_CPUS := runtime.NumCPU()
	runtime.GOMAXPROCS(NUM_CPUS)
	maxHandlers = NUM_CPUS
	if NUM_CPUS > 4 {
		// TODO: this needed? Or will running at a nice priority be sufficient?
		maxHandlers /= 2
	}

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

	case VERIFY:
		err := Verify(logs)
		if err != nil {
			out.Fatal(err)
		}

	// TODO - this is just for testing. remove later
	case "test-commit":
		err := TestCommit(logs)
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

	// find veb repo
	dir, err := cdBaseDir(logs)
	if err != nil {
		return err
	}

	// load the index
	index, err := veb.Load(logs)
	if err != nil {
		return fmt.Errorf("veb could not load index")
	}

	// check for changes
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(".", files, false)

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
		fmt.Println("  (use 'veb fix <file>' if a file has been corrupted in this repository)")
		fmt.Println("  (use 'veb push', 'veb pull', or 'veb sync' to commit changed/new files)")
	}
	fmt.Printf("\nsummary: %d new, %d changed\n", len(newFiles), len(changedFiles))

	timer.Stop()
	logs.Info.Println(OUT_FUNC, STATUS)
	logs.Info.Printf("%s (%d new, %d changed) took %v\n",
		STATUS, len(newFiles), len(changedFiles), timer.Duration())
	return nil
}

// Runs every file in index through hashing algorithm and compares the result
// against the xsum saved in the index.
// Does not verify new files.
func Verify(logs *veb.Logs) error {
	logs.Info.Println(IN_FUNC, VERIFY)
	var timer veb.Timer
	timer.Start()

	// find veb repo
	dir, err := cdBaseDir(logs)
	if err != nil {
		return err
	}

	// load the index
	index, err := veb.Load(logs)
	if err != nil {
		return fmt.Errorf("veb could not load index")
	}

	// start listener for user's quit signal
	quit := make(chan int)
	go func() {
		for {
			var input string
			fmt.Scan(&input)
			if input[0] == QUIT_RUNE {
				quit <- 1
			}
		}
	}()

	// print intro
	fmt.Println("veb repository at", dir, "\n")
	fmt.Println("Verifying file checksums against those stored in veb index...")
	fmt.Println("Note: new files (as shown by 'veb status') will not be checked.\n")

	// toss everything in index into input channel
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go func() {
		for _, f := range index.Files {
			files <- f
		}
		close(files)
	}()

	// start handler pool working on checking files
	changed := make(chan veb.IndexEntry, CHAN_SIZE)
	done := make(chan int, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		go verifyHandler(files, changed, done, logs)
	}

	// done listener signals quit when all handlers are done
	go func() {
		for i := 0; i < maxHandlers; i++ {
			<-done
		}
		close(changed)
		quit <- 1
	}()

	for {
		i := 0 // TODO testing...
		select {
		case <-quit:
			// We're done! Either by finishing or user interrupt.
			// TODO
			// print outro

			// info logs
			timer.Stop()
			logs.Info.Println(OUT_FUNC, VERIFY)
			logs.Info.Printf("%s (%d ok, %d changed, %d not checked) took %v\n",
				VERIFY, 0, 0, 0, timer.Duration()) // TODO: stats in there instead of zeros
			return nil

		case f := <-changed:
			// clear status line w/ carriage return & 80 spaces
			fmt.Println("\r                                                                                ")

			// TODO
			//  - will need a 'print changes' func.
			//  - ERROR!!!!1~! when something's xsum is different but stats are the same
			//  - WARN probably when stats are different but xsum's the same?
			fmt.Println(f)

			// fallthrough to print status line
			fmt.Printf("\n")
			// TODO: actual status line
			i++
			fmt.Printf("\r%4d: ", i)
		default:
			// TODO: actual status line
			i++
			fmt.Printf("\r%4d: ", i)
		}
	}

	// never should get here
	return nil
}

func TestCommit(logs *veb.Logs) error {
	logs.Info.Println(IN_FUNC, "test-commit")
	var timer veb.Timer
	timer.Start()

	// find veb repo
	dir, err := cdBaseDir(logs)
	if err != nil {
		return err
	}
	fmt.Println("veb repository at", dir, "\n")

	// load the index
	index, err := veb.Load(logs)
	if err != nil {
		return fmt.Errorf("veb could not load index")
	}

	// check for changes
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(".", files, false)
	
	// start handler pool working on files
	updates := make(chan veb.IndexEntry, CHAN_SIZE)
	done := make(chan int, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		go func() {
			for f := range files {
				// calculate checksum hash
				err = veb.Xsum(&f, logs)
				if err != nil {
					logs.Err.Println("checksum for verify failed:", err)
				}
				
				updates <- f
			}
			done <- 1
		}()
	}

	// done listener
	go func() {
		for i := 0; i < maxHandlers; i++ {
			<-done
		}
		close(updates)
	}()

	// update index
	for f := range updates {
		index.Update(&f)
	}

	// save index once everything's done
	index.Save()

	// info logs
	timer.Stop()
	logs.Info.Println(OUT_FUNC, "test-commit")
	logs.Info.Printf("%s took %v\n",
		"test-commit", timer.Duration())
	return nil
}


// finds veb META_FOLDER and changes to that directory's parent
// returns: 
//   (dir cd'd to, nil) on success
//   ("", err) on error
func cdBaseDir(logs *veb.Logs) (string, error) {
	const MAX_PARENTS = 15
	pwd, err := os.Getwd()
	if err != nil {
		logs.Err.Println(err)
		return "", fmt.Errorf("veb could not figure out where it is: %v", err)
	}

	// search down the present working dir for META_FOLDER && cd there
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
				return "", fmt.Errorf("veb could not find %v metafolder: %v", META_FOLDER, err)
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
		return "", fmt.Errorf("veb could not find %v metafolder at or %v folders below: %v",
			META_FOLDER, MAX_PARENTS, pwd)
	}

	return dir, nil
}

// calculates checksum and saves file stats
func verifyHandler(files, changed chan veb.IndexEntry, done chan int, logs *veb.Logs) {
	for f := range files {
		// save off old xsum for comparison
		oldXsum := f.Xsum

		// get file size & such
		err := veb.SetStats(&f)
		if err != nil {
			logs.Err.Println("couldn't get stats:", err)
		}

		// calculate checksum hash
		err = veb.Xsum(&f, logs)
		if err != nil {
			logs.Err.Println("checksum for verify failed:", err)
		}

		// see if it changed...
		if !bytes.Equal(f.Xsum, oldXsum) {
			changed <- f
		}
	}
	done <- 1
}
