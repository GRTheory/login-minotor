package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	file := "./test.txt"

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = w.Add(filepath.Dir(file))
	if err != nil {
		log.Fatal("error:", err)
	}
	defer w.Close()

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			fmt.Println("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok {
				return
			}
			fmt.Println(e, e.Name)
			if e.Name == file {
				fmt.Println("is same")
			}
		}
	}
}

// func main() {
// 	// Create new watcher.
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer watcher.Close()

// 	// Start listening for events.
// 	go func() {
// 		for {
// 			select {
// 			case event, ok := <-watcher.Events:
// 				if !ok {
// 					return
// 				}
// 				log.Println("event:", event)
// 				if event.Has(fsnotify.Write) {
// 					log.Println("modified file:", event.Name)
// 				}
// 			case err, ok := <-watcher.Errors:
// 				if !ok {
// 					return
// 				}
// 				log.Println("error:", err)
// 			}
// 		}
// 	}()

// 	// Add a path.
// 	err = watcher.Add("/tmp")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Block main goroutine forever.
// 	<-make(chan struct{})
// }
