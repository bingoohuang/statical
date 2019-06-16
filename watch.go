package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

func watchSrc(srcPath string) {
	// https://github.com/elliotforbes/go-webassembly-framework/blob/master/internal/commands/start.go
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	err = watcher.Add(srcPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("started to watch src", srcPath)

	for {
		select {
		case e := <-watcher.Events:
			fmt.Println("src event detected", e)
			switch {
			case e.Op&fsnotify.Write == fsnotify.Write, e.Op&fsnotify.Remove == fsnotify.Remove:
				fmt.Println("rebuild")
				statiq()
			default:
				fmt.Println("ignored")
			}
		case err := <-watcher.Errors:
			fmt.Println("Error: ", err)
		}
	}

}
