package main

import (
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func innerLoop(directory string, signalsChan chan os.Signal, inEventChan chan fsnotify.Event, inErrChan chan error) bool {
	// do the task spawning and scanning functionality
	log.Info("Running work function")
	sleepseconds := doWork()

	// setup a timer and wait until the next scan time
	log.Debugf("Setting up timer for %s", formatElapsedTime(sleepseconds))
	timer := time.NewTimer(time.Duration(sleepseconds) * time.Second)
	timerChan := timer.C

	// loop until we receive a signal
	mustContinue := true
	stillWaiting := true
	for stillWaiting {
		select {
		case <-timerChan:
			log.Info("Signalled via timer")
			stillWaiting = false

		case s := <-signalsChan:
			log.Debugf("Signalled via signal [%s]", s.String())
			stillWaiting = false

			// if it was a sigterm, then we clean up
			if s == syscall.SIGTERM || s == syscall.SIGINT {
				mustContinue = false
			}

		case e := <-inEventChan:
			log.Debugf("Inotify event: %s", e)

		case e := <-inErrChan:
			log.Debugf("Inotify error: %s", e)

		}
	}

	// make sure the timer has stopped/been consumed no matter what signal
	timer.Stop()
	return mustContinue
}

func foreverLoop(directory string) error {
	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT)

	tasksDirectory := path.Join(directory, "tasks.d")

	inEventChan := make(chan fsnotify.Event)
	inErrChan := make(chan error)

	log.Infof("Beginning inotify watcher, this might fail on an unsupported OS")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Could not create inotify watcher: %s", err)
	} else {
		defer watcher.Close()
		inEventChan = watcher.Events
		inErrChan = watcher.Errors

		log.Infof("Beginning to watch task directory %s", tasksDirectory)
		if err := watcher.Add(tasksDirectory); err != nil {
			return err
		}
	}

	for {
		if !innerLoop(directory, signalsChan, inEventChan, inErrChan) {
			break
		}
	}

	log.Info("Ending main loop.")
	return nil
}
