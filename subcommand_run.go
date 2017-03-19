package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/AstromechZA/ticktickd/pidfile"
	"github.com/tucnak/climax"
	"gopkg.in/natefinch/lumberjack.v2"

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
	}
	return nil
}

func checkTickTickDirectory(directory string) error {
	log.Printf("Checking ticktickd directory %s", directory)
	if err := checkDirectory(directory); err != nil {
		return err
	} else if unix.Access(directory, unix.W_OK) != nil {
		return fmt.Errorf("Directory path '%s' is not writeable", directory)
	}

	tasksDir := path.Join(directory, "tasks.d")
	log.Printf("Checking tasks.d directory %s", tasksDir)
	if err := checkDirectory(tasksDir); err != nil {
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
	logDir := path.Join(directory, "ticktickd.log")
	log.Printf("Beginning logging to %s", logDir)
	log.SetOutput(&lumberjack.Logger{
		Filename:   logDir,
		MaxSize:    128, // megabytes
		MaxBackups: 5,
		MaxAge:     90, //days
	})

	return foreverLoop(directory, mustWatch)
}
