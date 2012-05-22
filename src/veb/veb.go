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
	"io"
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
	COMMIT = "commit"

	// misc
	QUIT_RUNE = 'q'
	CHAN_SIZE = 1000
	INDENT_F = " " // use with Println == 2 spaces
	INDENT_I = "      -"

	// TODO: move somewhere else. Index uses thes too.
	META_FOLDER = ".veb"
	INDEX_FILE  = "index" // inside of META_FOLDER only
	XSUMS_FILE  = "xsums" // inside of META_FOLDER only
	LOG_FILE    = "log.txt"
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

	// init is a bit different (no pre-existing index),
	// so take care of it here instead of inside switch
	if flag.Args()[0] == INIT {
		err := Init()
		if err != nil {
			out.Fatal(err)
		}
		pwd, _ := os.Getwd()
		out.Println("Initialized empty veb repository at", pwd)
		return // done
	}

	// find veb repo
	root, err := cdBaseDir()
	if err != nil {
		fmt.Println(err, "\n")
		out.Fatal("Use 'veb init' to create this veb repository.")
	}

	// make the logger
	logf, err := os.OpenFile(path.Join(root, META_FOLDER, LOG_FILE),
		os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	log := veb.NewLog(log.New(logf, "", log.LstdFlags|log.Lshortfile))

	// load the index
	index, err := veb.Load(root, log)
	if err != nil {
		out.Fatal("veb could not load index")
	}

	// print intro
	fmt.Println("veb repository at", root, "\n")

	// act on command
	switch flag.Args()[0] {
	case STATUS:
		err = Status(index, log)
		if err != nil {
			out.Fatal(err)
		}

	case VERIFY:
		err = Verify(index, log)
		if err != nil {
			out.Fatal(err)
		}

	case COMMIT:
		err = Commit(index, log)
		if err != nil {
			out.Fatal(err)
		}

	case REMOTE:
		if len(flag.Args()) < 2 {
			out.Fatal(REMOTE, " needs a path to the backup repository",
				"\n  e.g. 'veb remote ~/backups/music'")
		}
		err = Remote(index, flag.Args()[1], log)
		if err != nil {
			out.Fatal(err)
		}

	case PUSH:
		err = Push(index, log)
		if err != nil {
			out.Fatal(err)
		}

	case FIX:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case PULL:
		// TODO: implement
		out.Fatal("this command is not yet implemented")
	case SYNC:
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

	log.Info().Println("done\n\n")
}

// creates veb's META_FOLDER in current directory and empty veb metadata 
// files inside META_FOLDER
func Init() error {
	// create veb dir
	err := os.Mkdir(path.Join(META_FOLDER), 0755)
	if err != nil {
		return fmt.Errorf("veb could not create metadata directory: %v", err)
	}

	// create index file
	indexf, err := os.Create(path.Join(META_FOLDER, INDEX_FILE))
	if err != nil {
		return fmt.Errorf("veb could not create metadata index file: %v", err)
	}
	indexf.Close()

	// create xsums file
	xsums, err := os.Create(path.Join(META_FOLDER, XSUMS_FILE))
	if err != nil {
		return fmt.Errorf("veb could not create metadata xsums file: %v", err)
	}
	xsums.Close()

	// create & save empty index
	index := veb.New(crypto.SHA1, ".")
	err = index.Save()
	if err != nil {
		return err
	}

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
func Status(index *veb.Index, log *veb.Log) error {
	defer log.Un(log.Trace(STATUS))
	var timer veb.Timer
	timer.Start()

	// check for changes
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(files)

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
				log.Err().Println(err)
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
				log.Err().Println(err)
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
					sizeChange = -sizeChange // absolute value
				}

				sanity := false

				// print size change
				if sizeChange != 0 {
					fmt.Printf("%s filesize %s %s (%s -> %s)\n",
						INDENT_I, direction, sizeChange, prevSize, curSize)
					sanity = true
				}

				// print mtime
				if index.Files[f].ModTime != fi.ModTime() {
					fmt.Printf("%s modified on (%v)\n", INDENT_I, fi.ModTime())
					sanity = true
				}

				// print mode
				if index.Files[f].Mode != fi.Mode() {
					fmt.Printf("%s file mode changed (%v -> %v)\n",
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
	timer.Stop()
	fmt.Printf("\nsummary: %d new, %d changed (%v)\n", 
		len(newFiles), len(changedFiles), timer.Duration())

	log.Info().Printf("%s (%d new, %d changed) took %v\n",
		STATUS, len(newFiles), len(changedFiles), timer.Duration())
	return nil
}

// Runs every file in index through hashing algorithm and compares the result
// against the xsum saved in the index.
// Does not verify new files.
func Verify(index *veb.Index, log *veb.Log) error {
	defer log.Un(log.Trace(VERIFY))
	var timer veb.Timer
	timer.Start()

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
	fmt.Println("Verifying file checksums against those stored in veb index...")
	fmt.Println("Note: new files (as shown by 'veb status') will not be checked.\n")

	// bail early for empty index
	if len(index.Files) == 0 {
		fmt.Println("No files in veb index. Nothing to verify.")
		return nil
	}

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
		go verifyHandler(index.Root, files, changed, done, log)
	}

	// done listener signals quit when all handlers are done
	go func() {
		for i := 0; i < maxHandlers; i++ {
			<-done
		}
		quit <- 1
	}()

	first := true
	totalFiles := len(index.Files)
	changedFiles := 0
	scannedFiles := 0
verify_receive_loop:
	for {
		select {
		case <-quit:
			// We're done! Either by finishing or user interrupt.
			break verify_receive_loop

		case f := <-changed:
			// clear status line w/ carriage return & 80 spaces
			fmt.Println("\r                                                                                ")

			// print header once first file is encountered
			if first {
				fmt.Println("----------------------")
				fmt.Println("Files with new hashes:")
				fmt.Println("----------------------")
				first = false
			}

			// TODO
			//  - move all this to a 'print changed' function?

			// print file name
			fmt.Println(INDENT_F, f.Path)

			// figure out filesize
			curSize := ByteSize(f.Size)
			prevSize := ByteSize(index.Files[f.Path].Size)
			sizeChange := curSize - prevSize
			direction := "increased"
			if sizeChange < 0 {
				direction = "decreased"
				sizeChange = -sizeChange // absolute value
			}
			
			sanity := false
			
			// print size change
			if sizeChange != 0 {
				fmt.Printf("%s filesize %s %s (%s -> %s)\n",
					INDENT_I, direction, sizeChange, prevSize, curSize)
				sanity = true
			}
			
			// print mtime
			if index.Files[f.Path].ModTime != f.ModTime {
				fmt.Printf("%s modified on (%v)\n", INDENT_I, f.ModTime)
				sanity = true
			}
			
			// print mode
			if index.Files[f.Path].Mode != f.Mode {
				fmt.Printf("%s file mode changed (%v -> %v)\n",
					INDENT_I, index.Files[f.Path].Mode, f.Mode)
				sanity = true
			}
			
			// sanity check & snark
			if !sanity {
				fmt.Printf("%s ...well /something/ changed. Dunno what. *shrugs*\n", INDENT_I)
			}
			
			// print xsums
			// TODO: dynamic hash name instead of hard 'SHA1'
			fmt.Printf("%s previous SHA1: %x\n", INDENT_I, index.Files[f.Path].Xsum)
			fmt.Printf("%s current  SHA1: %x\n", INDENT_I, f.Xsum)

			fmt.Printf("\n")

			// status line
			changedFiles++
			scannedFiles = totalFiles - len(files)
			fmt.Printf("\rscanned: %6d of %6d files (%d changed) (type 'q' to quit): ",
				scannedFiles, totalFiles, changedFiles)

		default:
			// status line
			scannedFiles = totalFiles - len(files)
			fmt.Printf("\rscanned: %6d of %6d files (%d changed) (type 'q' to quit): ",
				scannedFiles, totalFiles, changedFiles)
		}
	}

	notChecked := totalFiles - scannedFiles
	okFiles := totalFiles - changedFiles - notChecked
	
	// print outro
	timer.Stop()
	fmt.Println("\n\nMAKE SURE CHANGED FILES ARE THINGS YOU'VE ACTUALLY CHANGED")
	fmt.Println("  (use 'veb fix <file>' if a file has been corrupted in this repository)")
	fmt.Println("  (use 'veb push', 'veb pull', or 'veb sync' to commit changed/new files)")
	fmt.Printf("\nsummary: %d ok, %d changed, %d not checked in %v\n",
		okFiles, changedFiles, notChecked, timer.Duration())
	
	// info log
	log.Info().Printf("%s (%d ok, %d changed, %d not checked) took %v\n",
		VERIFY, okFiles, changedFiles, notChecked, timer.Duration())
	return nil
}

// Saves all updated/new files to index, so they are available for push/pull
func Commit(index *veb.Index, log *veb.Log) error {
	defer log.Un(log.Trace(COMMIT))
	var timer veb.Timer
	timer.Start()

	// check for changes
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	go index.Check(files)
	
	// start handler pool working on files
	updates := make(chan veb.IndexEntry, CHAN_SIZE)
	done := make(chan int, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		go func() {
			for f := range files {
				// calculate checksum hash
				err := veb.Xsum(&f, log)
				if err != nil {
					log.Err().Println("checksum for verify failed:", err)
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
	var retVal error = nil
	numCommits := 0
	numErrors := 0
	for f := range updates {
		err := index.Update(&f)
		if err != nil {
			log.Err().Println("index update failed:", err)
			fmt.Println("Error: Couldn't commit", f.Path, ":", err)
			retVal = fmt.Errorf("veb commit failed")
			numErrors++
		} else {
			fmt.Println(INDENT_F, "committed", f.Path)
			numCommits++
		}
	}

	// save index once everything's done
	index.Save()

	// info 
	timer.Stop()
	fmt.Println("summary:", numCommits, "commits,", numErrors, 
		"errors in", timer.Duration())
	log.Info().Printf("%s (%d commits, %d errors) took %v\n",
		COMMIT, numCommits, numErrors, timer.Duration())
	return retVal
}

// sets the remote repository
func Remote(index *veb.Index, remote string, log *veb.Log) error {
	defer log.Un(log.Trace(REMOTE))
	var timer veb.Timer
	timer.Start()

	// make remote an absolute path
	if !path.IsAbs(remote) {
		remote = path.Join(index.Root, remote)
	}	

	// check to see if remote exists
	fi, err := os.Stat(remote)
	if err != nil {
		if os.IsNotExist(err) {
			log.Err().Println(err)
			return fmt.Errorf("veb remote dir does not exist: %v", err)
		} else {
			log.Err().Println(err)
			return err
		}
	} else if !fi.IsDir() {
		// ain't a directory
		log.Err().Println(remote, "isn't a directory")
		return fmt.Errorf("veb remote must be a folder: %s is not a folder", remote)
	}

	// check to see if it's a veb repo
	remoteRepo := path.Join(remote, META_FOLDER)
	fi, err = os.Stat(remoteRepo)
	if err != nil {
		if os.IsNotExist(err) {
			log.Err().Println(err)
			fmt.Println("veb remote needs to be initialized as a veb repository",
				"\n  (use 'veb init' in remote dir)")
			return err
		} else {
			log.Err().Println(err)
			return err
		}
	} else if !fi.IsDir() {
		// ain't a directory
		log.Err().Println(remoteRepo, "isn't a directory")
		fmt.Println("veb remote needs", remoteRepo, "to be a folder",
			"\nDelete or rename that file and run 'veb init' from", remote, "\n")
		return fmt.Errorf("%s isn't a directory", remoteRepo)
	}

	// set remote
	index.Remote = remote
	index.Save()
	fmt.Println("veb added", remote, "as the remote")

	// info log
	timer.Stop()
	log.Info().Printf("%s took %v\n",
		REMOTE, timer.Duration())
	return nil
}

// gather info about new/changed local files, then copies these files
// to the remote location if remote doesn't have same checksum.
// updates metadata both locally and remotely.
func Push(local *veb.Index, log *veb.Log) error {
	defer log.Un(log.Trace(PUSH))
	var timer veb.Timer
	timer.Start()

	// open remote's index
	if local.Remote == "" {
		return fmt.Errorf("No remote veb repository. Use 'veb remote' to set one.")
	}
	remote, err := veb.Load(local.Remote, log) // TODO: have log indicate local vs remote
	if err != nil {
		return fmt.Errorf("veb could not load remote index: %v", err)
	}

	// get new/changed files for local & remote
	// we'll ignore these, as they haven't been committed
	locIgnore := make(chan veb.IndexEntry, CHAN_SIZE)
	remIgnore := make(chan veb.IndexEntry, CHAN_SIZE)
	go local.Check(locIgnore)
	go remote.Check(remIgnore)

	// notify user of ignored files
	cmt := true
	cmtMsg := func() {
		if cmt {
			fmt.Println("use 'veb status' to check new/changed files")
			fmt.Println("use 'veb commit' to add new/changed files to repository")
			cmt = false
		}		
	}
	first := true
	locFilter := make(map[string]bool)
	remFilter := make(map[string]bool)
	for f := range locIgnore {
		if first {
			cmtMsg()
			fmt.Println("\n--------------------")
			fmt.Println("LOCAL ignored files:")
			fmt.Println("--------------------")
			first = false
		}
		fmt.Println(INDENT_F, f.Path)
		locFilter[f.Path] = true
	}
	first = true
	for f := range remIgnore {
		if first {
			cmtMsg()
			fmt.Println("\n---------------------")
			fmt.Println("REMOTE ignored files:")
			fmt.Println("---------------------")
			first = false
		}
		fmt.Println(INDENT_F, f.Path)
		remFilter[f.Path] = true
	}

	// make list of files to check
	files := make(chan veb.IndexEntry, CHAN_SIZE)
	numIgnored := 0
	go func() {
		for p, f := range local.Files {
			// ignore if it's one of the new/changed files
			_, skipL := locFilter[p]
			_, skipR := remFilter[p]
			if skipL || skipR {
				numIgnored++
			} else {
				files <- f
			}
		}
		close(files)
	}()

	// send files to remote if xsums differ
	done := make(chan int, maxHandlers)
	updates := make(chan veb.IndexEntry, CHAN_SIZE)
	errored := make(chan veb.IndexEntry, CHAN_SIZE)
	numErrored := 0
	numPushed  := 0
	numNoChange := 0
	for i := 0; i < maxHandlers; i++ {
		go func() {
			for f := range files {
				// compare checksum hashes
				_, ok := remote.Files[f.Path] // does file exist in remote yet?
				if ! ok || !bytes.Equal(f.Xsum, remote.Files[f.Path].Xsum) {
					// TODO: verify f.Xsum == local file's actual xsum
					//  - don't want corrupted files getting across.

					err := pushFile(local.Root, remote.Root, f, log)
					if err != nil {
						// notify of error, but continue with rest of files
						// TODO: get the error out too
						errored <- f
						numErrored++
					} else {
						// save entry so index can be updated
						updates <- f
						numPushed++
					}
				} else {
					numNoChange++
				}
			}
			done <- 1
		}()
	}

	// listeners
	var retVal error = nil
	numListeners := 3
	quit := make(chan int, numListeners)
	go func() {
		for i := 0; i < maxHandlers; i++ {
			<-done
		}
		close(updates)
		close(errored)
		quit <- 1
	}()
	go func() {
		for f := range errored {
			// clear status line w/ 80 spaces & carriage return
			fmt.Printf("\r%sError: could not push to remote: %s\n",
				"                                                                                \r",
				f.Path)

			if retVal == nil {
				retVal = fmt.Errorf("error transferring files to remote")
			}
		}
		quit <- 1
	}()
	go func () {
		first := true
		for f := range updates {
			if first {
				fmt.Printf("%s%s\n%s\n%s\n",
					"\r                                                                                \r",
					"-------------",
					"Pushed files:",
					"-------------")
				first = false
			}

			// clear status line w/ 80 spaces & carriage return
			fmt.Printf("\r%s%s %s\n",
				"                                                                                \r",
				INDENT_F, f.Path)
			remote.Update(&f)
		}
		quit <- 1
	}()

	// print status while waiting for everyone to finish
	for len(quit) < numListeners {
		fmt.Printf("\rstatus: %4d ignored, %4d errors, %4d pushed, %4d unchanged",
			numIgnored, numErrored, numPushed, numNoChange)
	}	

	// save remote index's updates
	remote.Save()

	// print outro
	timer.Stop()
	fmt.Println("\r                                                                                ")
	fmt.Printf("status: %4d ignored, %4d errors, %4d pushed, %4d unchanged in %v\n",
		numIgnored, numErrored, numPushed, numNoChange, timer.Duration())
	
	// info log
	timer.Stop()
	log.Info().Printf("%s (%d ignored, %d errors, %d pushed, %d unchanged) took %v\n",
		PUSH, numIgnored, numErrored, numPushed, numNoChange, timer.Duration())
	return retVal
}

// finds veb META_FOLDER and changes to that directory's parent
// returns: 
//   (dir cd'd to, nil) on success
//   ("", err) on error
func cdBaseDir() (string, error) {
	const MAX_PARENTS = 15
	pwd, err := os.Getwd()
	if err != nil {
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
				return "", fmt.Errorf("veb could not find %v metafolder: %v", META_FOLDER, err)
			}
		} else if fi.IsDir() {
			// dir is now at base directory
			found = true
			err = os.Chdir(dir)
			if err != nil {
				return "", fmt.Errorf("veb could not cd to base directory (%v): %v", dir, err)
			}
			break
		}

		// if not found, check parent dir next time
		dir = path.Dir(dir)
	}

	if !found {
		return "", fmt.Errorf("veb could not find %v metafolder at or (up to %v folders) below: %v",
			META_FOLDER, MAX_PARENTS, pwd)
	}

	return dir, nil
}

// calculates checksum and saves file stats
func verifyHandler(root string, files, changed chan veb.IndexEntry, done chan int, log *veb.Log) {
	for f := range files {
		// save off old xsum for comparison
		oldXsum := f.Xsum

		// get file size & such
		err := veb.SetStats(root, &f)
		if err != nil {
			log.Err().Println("couldn't get stats:", err)
		}

		// calculate checksum hash
		err = veb.Xsum(&f, log)
		if err != nil {
			log.Err().Println("checksum for verify failed:", err)
		}

		// see if it changed...
		if !bytes.Equal(f.Xsum, oldXsum) {
			changed <- f
		}
	}
	done <- 1
}

// pushes new file to remote repository
// TODO: Don't use Copy. Use rsync. rsync -qa
func pushFile(localRoot, remoteRoot string, entry veb.IndexEntry, log *veb.Log) error {
	// open local file
	local, err := os.Open(path.Join(localRoot, entry.Path))
	if err != nil {
		log.Err().Println(err)
		return err
	}
	defer local.Close()

	// make remote dirs, if they don't exist
	err = os.MkdirAll(path.Dir(path.Join(remoteRoot, entry.Path)), 0755)
	if err != nil {
		log.Err().Println(err)
		return err
	}

	// open remote file
	remote, err := os.OpenFile(path.Join(remoteRoot, entry.Path),
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE, entry.Mode)
	if err != nil {
		log.Err().Println(err)
		return err
	}
	defer remote.Close()

	// send it!
	_, err = io.Copy(remote, local)
	if err != nil {
		log.Err().Println(err)
		return err
	}

	time.Sleep(time.Second)
	return err
}
