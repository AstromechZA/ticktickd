package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func innerLoop(signalsChan chan os.Signal) {
	// do the task spawning and scanning functionality
	log.Info("Running work function..")
	sleepseconds := doWork()

	// setup a timer and wait until the next scan time
	log.Infof("Setting up timer for %s..", formatElapsedTime(sleepseconds))
	timer := time.NewTimer(time.Duration(sleepseconds) * time.Second)
	timerChan := timer.C

	// loop until we receive a signal
	stillWaiting := true
	for stillWaiting {
		select {
		case <-timerChan:
			log.Info("Signalled via timer")
			stillWaiting = false
		case s := <-signalsChan:
			log.Infof("Signalled via signal [%s]", s.String())
			stillWaiting = false
		}
	}

	// make sure the timer has stopped/been consumed no matter what signal
	timer.Stop()
}

func foreverLoop() error {
	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, syscall.SIGUSR1)

	for {
		innerLoop(signalsChan)
	}
}
