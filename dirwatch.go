package main

import (
	"exp/inotify"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

type Dirwatch interface {
	IsValid(path, dir, file string) bool
	IsModified(path, dir, file string) bool
	OnDelete(path, dir, file string)
	OnChange(path, dir, file string)
}

type baseDirwatch struct {
	Paths   []string
	IFlags  int
	watcher *inotify.Watcher
}

func (d *baseDirwatch) IsValid(path, dir, file string) bool {
	log.Println("dirwatch: you should override IsValid")
	return false
}

func (d *baseDirwatch) IsModified(path, dir, file string) bool {
	log.Println("dirwatch: you should override IsModified")
	return false
}

func (d *baseDirwatch) OnDelete(path, dir, file string) {
	log.Println("dirwatch: you should override OnDelete")
}

func (d *baseDirwatch) OnChange(path, dir, file string) {
	log.Println("dirwatch: you should override OnChange")
}

func isMedia(path string) bool {
	switch {
	case strings.HasSuffix(path, ".mp3"):
		return true
	case strings.HasSuffix(path, ".m4a"):
		return true
	case strings.HasSuffix(path, ".ogg"):
		return true
	case strings.HasSuffix(path, ".flac"):
		return true
	}
	return false
}

func processDir(base, dir string, dirs, mediaFiles chan string) (err error) {
	log.Printf("adding %s\n", dir)
	var fullPath string
	// special case the top level dir
	if base == dir {
		fullPath = base
		dir = ""
	} else {
		fullPath = path.Join(base, dir)
	}
	info, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return
	}
	for _, fi := range info {
		name := fi.Name()
		if fi.IsDir() {
			dirs <- path.Join(dir, name)
		} else if isMedia(name) {
			mediaFiles <- path.Join(dir, name)
		}
	}
	return
}

func initWatch(base string, dirs, mediaFiles chan string) (err error) {
outer:
	for {
		select {
		case dir := <-dirs:
			if err = processDir(base, dir, dirs, mediaFiles); err != nil {
				return
			}
		default:
			break outer
		}
	}
	return nil
}

func NewDirwatch(path string) (result Dirwatch, err error) {
	d := new(baseDirwatch)

	if d.watcher, err = inotify.NewWatcher(); err != nil {
		return
	}

	mediaFiles := make(chan string, 10240)
	dirs := make(chan string, 1024)
	dirs <- path

	if err = initWatch(path, dirs, mediaFiles); err != nil {
		return
	}
	log.Printf("len files: %d\n", len(mediaFiles))

	return d, nil
}
