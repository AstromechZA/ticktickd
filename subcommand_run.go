package main

import (
	"fmt"
	"os"
	"time"

	"github.com/AstromechZA/ticktickd/pidfile"

	"golang.org/x/sys/unix"
)

func checkDirectory(directory string) error {
	if st, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Directory path '%s' does not exist", directory)
		}
		return fmt.Errorf("Failed to read directory '%s': %s", directory, err.Error())
	} else if !st.IsDir() {
		return fmt.Errorf("Directory path '%s' is not a directory", directory)
	} else if unix.Access(directory, unix.W_OK) != nil {
		return fmt.Errorf("Directory path '%s' is not writeable", directory)
	}
	return nil
}

func subcommandRun(directory string) error {

	// check that the directory exists
	if err := checkDirectory(directory); err != nil {
		return fmt.Errorf("Cannot start: %s", err.Error())
	}
	// check and write the pidfile
	pf, err := pidfile.NewPidfileAndWrite(directory, "ticktickd.pid")
	if err != nil {
		return fmt.Errorf("pidfile error: %s", err)
	}
	defer pf.Remove()

	time.Sleep(time.Minute)

	return nil
}
