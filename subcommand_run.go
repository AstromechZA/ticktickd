package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/AstromechZA/ticktickd/pidfile"
	"github.com/tucnak/climax"
	"golang.org/x/sys/unix"
	"gopkg.in/natefinch/lumberjack.v2"
)

func checkDirectory(directory string) error {
	log.Printf("Checking directory %s", directory)
	if st, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Directory path '%s' does not exist", directory)
		}
		return fmt.Errorf("Failed to read directory '%s': %s", directory, err.Error())
	} else if !st.IsDir() {
		return fmt.Errorf("Directory path '%s' is not a directory", directory)
	}
	return nil
}

func ensureDirectory(directory string) error {
	log.Printf("Ensuring directory %s", directory)
	if st, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating directory since it doesn't exist")
			if err := os.Mkdir(directory, 0755); err != nil {
				return fmt.Errorf("Failed to create directory '%s': %s", directory, err)
			}
			return nil
		}
		return fmt.Errorf("Failed to read directory '%s': %s", directory, err.Error())
	} else if !st.IsDir() {
		return fmt.Errorf("Directory path '%s' is not a directory", directory)
	}
	return nil
}

func checkTickTickDirectory(directory string) error {
	if err := checkDirectory(directory); err != nil {
		return err
	} else if unix.Access(directory, unix.W_OK) != nil {
		return fmt.Errorf("Directory path '%s' is not writeable", directory)
	}

	tasksDir := path.Join(directory, "tasks.d")
	if err := ensureDirectory(tasksDir); err != nil {
		return err
	} else if unix.Access(tasksDir, unix.X_OK|unix.R_OK) != nil {
		return fmt.Errorf("Directory path '%s' is not readable", tasksDir)
	}
	return nil
}

func subcommandRun(ctx climax.Context) error {
	// parse args and things
	directory := DefaultDirectory
	if d, ok := ctx.Get("directory"); ok {
		directory = d
	}
	mustWatch := !ctx.Is("disablewatch")

	// check that the directory exists
	if err := checkTickTickDirectory(directory); err != nil {
		return fmt.Errorf("Cannot start: %s", err.Error())
	}

	// check and write the pidfile
	log.Printf("Checking pidfile")
	pf, err := pidfile.NewPidfileAndWrite(directory, "ticktickd.pid")
	if err != nil {
		return fmt.Errorf("pidfile error: %s", err)
	}
	log.Printf("Wrote %s", pf.Path())
	defer pf.Remove()

	// check and setup rotating logs
	if !ctx.Is("nologfile") {
		logDir := path.Join(directory, "ticktickd.log")
		log.Printf("Switching to logfile %s", logDir)
		log.SetOutput(&lumberjack.Logger{
			Filename:   logDir,
			MaxSize:    128, // megabytes
			MaxBackups: 5,
			MaxAge:     90, //days
		})
	}
	return foreverLoop(directory, mustWatch)
}
