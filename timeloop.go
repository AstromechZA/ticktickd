package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func foreverLoop(directory string, watchTasksDir bool) error {
	tasksDirectory := path.Join(directory, "tasks.d")
	overall := make(chan bool)

	timer := time.NewTimer(time.Hour * 24)
	timerChan := timer.C
	go func() {
		for {
			<-timerChan
			log.Printf("Signalled via timer")
			overall <- true
		}
	}()

	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for s := range signalsChan {
			log.Printf("Signalled via signal [%s]", s.String())
			// if it was a sigterm, then we clean up
			if s == syscall.SIGTERM || s == syscall.SIGINT {
				overall <- false
			} else if s == syscall.SIGUSR1 {
				overall <- true
			}
		}
	}()

	inEventChan := make(chan fsnotify.Event)
	if watchTasksDir {
		log.Printf("Beginning inotify watcher, this might fail on an unsupported OS")
		log.Printf("This may cause increased cpu usage on some operating systems, disable it if needed using the cli")
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("Could not create inotify watcher: %s", err)
		} else {
			defer watcher.Close()
			inEventChan = watcher.Events

			log.Printf("Beginning to watch task directory %s", tasksDirectory)
			if err := watcher.Add(tasksDirectory); err != nil {
				return err
			}
		}
		go func() {
			for e := range inEventChan {
				if e.Op.String() != "" {
					log.Printf("Inotify event: %s", e)
					overall <- true
				}
			}
		}()
	}

	for {
		sleepDuration := doWork(directory)
		timer.Reset(sleepDuration)
		if c := <-overall; !c {
			break
		}
	}

	log.Printf("Ending main loop.")
	return nil
}
