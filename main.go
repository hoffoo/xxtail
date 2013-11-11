package main

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"io"
	"os"
	//"os/signal"
	//"syscall"
	"flag"
	"path"
	"path/filepath"
	"sync"
)

var mu *sync.RWMutex
var watcher *fsnotify.Watcher
var watching map[string]*os.File
var cwd string

var (
	recursive bool // recurse into directories
	watchAll  bool // watch hidden files
)

// function pased to os.Walk
func Tail(path string, info os.FileInfo, err error) error {

	// FIXME this wont skip hidden files on windows
	// not sure if people even use them
	if watchAll == false && info.Name()[0:1] == "." {
		return filepath.SkipDir
	}

	watcher.Watch(path)
	addWatch(path, true)
	fmt.Println(path)

	return nil
}

func update(in io.Reader) {
	b := make([]byte, 4096)
	for {
		n, err := in.Read(b)
		os.Stdout.Write(b[:n])

		if err != nil {
			break
		}
	}
}

func addWatch(path string, seekEnd bool) error {

	mu.Lock()
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	watching[path] = fd
	mu.Unlock()

	if seekEnd {
		fd.Seek(0, os.SEEK_END)
	}

	return nil
}

func watch() {

	var event *fsnotify.FileEvent
	var open bool

	for {
		if event, open = <-watcher.Event; open == false {
			break
		} else if event.IsModify() {
			go fileModified(event)
		} else if event.IsDelete() {
			out("DELETED", event.Name)
		} else if event.IsCreate() {
			go fileCreated(event)
		} else if event.IsRename() {
		}
	}
}

var lastformat string

func out(action, file string) {
	format := fmt.Sprintf("%s <== %s\n", file, action)
	if format == lastformat {
		return
	}

	lastformat = format
	fmt.Print(format)
}

func fileModified(event *fsnotify.FileEvent) {

	mu.RLock()
	fd, exists := watching[event.Name]
	mu.RUnlock()

	if exists == false {
		err := addWatch(event.Name, false)
		if err != nil {
			fmt.Printf("couldnt open file: %s", event.Name)
			return
		}
		fd = watching[event.Name]
		fd.Seek(0, os.SEEK_END)
	} else {
		out("MODIFIED", event.Name)
		update(fd)
	}
}

func fileCreated(event *fsnotify.FileEvent) {

	mu.RLock()
	fd, exists := watching[event.Name]
	mu.RUnlock()

	if exists == false {
		err := addWatch(event.Name, false)
		if err != nil {
			fmt.Printf("couldnt open file: %s", event.Name)
			return
		}
		out("CREATE", event.Name)
	} else {
		fd.Seek(0, 0)
		out("TRUNCATE", event.Name)
	}
}

func main() {

	flag.BoolVar(&recursive, "R", false, "tail failes in subfolders")
	flag.BoolVar(&watchAll, "a", false, "watch hidden directories")
	flag.Parse()
	target := flag.Arg(0)

	var err error

	cwd, err = os.Getwd()
	if err != nil {
		fmt.Println("couldnt get cwd")
		return
	}

	// if no args its cwd, otherwise the directory/file to tail
	if target == "" {
		target = cwd
	} else {
		_, staterr := os.Stat(path.Join(cwd, target))

		if staterr != nil {
			fmt.Println("couldnt stat " + target)
			return
		}
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("couldnt open watcher")
		return
	}

	watching = make(map[string]*os.File)
	mu = &sync.RWMutex{}

	go filepath.Walk(target, Tail)
	watch()
}
