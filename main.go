package main

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"io"
	"os"
	//"os/signal"
	//"syscall"
	"flag"
	"log"
	"path/filepath"
)

var watcher *fsnotify.Watcher
var watching map[string]*os.File
var cwd string
var recursive bool

var (
	modifyWatch chan *fsnotify.Event
	deleteWatch chan *fsnotify.Event
	createWatch chan *fsnotify.Event
	renameWatch chan *fsnotify.Event
)

func Tail(path string, info os.FileInfo, err error) error {
	watcher.Watch(cwd + "/" + path)
	log.Printf("==> %s <==", cwd+"/"+path)

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

func watch() {

	var exists bool // if the file is in our watching map
	var file *os.File
	var err error

	for {

		if event, chanOpen := <-watcher.Event

		if event.IsModify() {
			go update(file)
		} else if event.IsCreate() {

			if file, exists = watching[event.Name]; exists == false {
				file, err = os.Open(event.Name)
				if err != nil {
					panic(err)
				}
				watcher.Watch(event.Name)
				watching[event.Name] = file
			}

			file, err = os.Open(event.Name)
			if err != nil {
				panic(err)
			}
			fmt.Printf("==> %s <==\n", event.Name)
			watching[event.Name] = file
			go update(file)
		}
	}
}

func main() {

	flag.BoolVar(&recursive, "R", false, "tail failes in subfolders")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("couldnt get cwd")
		return
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("couldnt open watcher")
		return
	}

	watching = make(map[string]*os.File)

	//	sigchan := make(chan os.Signal)
	//	signal.Notify(sigchan, syscall.SIGINT)
	//
	//	go func() {
	//		for {
	//			<-sigchan
	//			for file, _ := range watching {
	//				fmt.Printf("%s\n", file)
	//			}
	//			watcher.Close()
	//		}
	//	}()

	go filepath.Walk(cwd, Tail)

	watch()

}
