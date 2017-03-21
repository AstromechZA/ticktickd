package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"syscall"

	"github.com/AstromechZA/ticktickd/pidfile"
	"github.com/tucnak/climax"
)

func printProcessInformation(directory string) {
	fmt.Println("Process:")

	// setup pidfile
	pf, err := pidfile.NewPidfile(directory, "ticktickd.pid")
	if err != nil {
		fmt.Println("  can't tell if ticktickd is running (pidfile could not be read)")
		return
	}

	// read it
	pid, err := pf.Read()
	if err != nil {
		fmt.Println("  can't tell if ticktickd is running (pidfile could not be read)")
		return
	}
	if pid == pidfile.MissingPidFile {
		fmt.Println("  ticktickd is not running (no pidfile)")
		return
	} else if pid < 2 {
		fmt.Println("  can't tell if ticktickd is running (invalid pid in pidfile)")
		return
	}

	// find and signal the process
	proc, _ := os.FindProcess(pid)
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		if err == syscall.ESRCH {
			fmt.Printf("  ticktickd is not running (pid %d is not running)\n", pid)
		} else {
			fmt.Printf("  can't tell if ticktickd is running (not allowed to query %d)\n", pid)
		}
		return
	}
	fmt.Printf("  ticktickd is running with pid %d\n", pid)
	fmt.Println()
}

func printTaskInformation(directory string) {
	fmt.Println("Tasks:")

	tasksDir := path.Join(directory, "tasks.d")
	if err := checkDirectory(tasksDir); err != nil {
		fmt.Printf("  tasks directory %s does not exist", tasksDir)
		return
	}

	taskDefs, failures, err := LoadTaskDefinitions(tasksDir)
	if err != nil {
		fmt.Printf("  could not load task definitions from %s: %s\n", tasksDir, err)
		return
	}

	db, _ := InitTimeDB(directory)
	defer db.Close()

	currentTime := time.Now()
	for _, td := range taskDefs {
		fmt.Printf("  %s:\n", td.FileName)
		err := td.Validate()
		if err != nil {
			fmt.Printf("    %s\n", err)
			continue
		}

		r, _ := td.GetRule()
		fmt.Printf("    name: %s\n", td.Name)
		fmt.Printf("    rule: %s\n", td.Rule)
		lr := GetLastRunTime(db, &td)
		if lr.IsZero() {
			fmt.Printf("    last run: never\n")
		} else {
			fmt.Printf("    last run: %s\n", lr)
		}

		nr := r.NextAfter(currentTime)
		tu := r.UntilNext(currentTime)

		fmt.Printf("    next run: %s (in %s)\n", nr, tu)
	}

	for fn, e := range failures {
		fmt.Printf("  %s:\n", fn)
		fmt.Printf("    %s\n", e)
	}

	if len(taskDefs) == 0 && len(failures) == 0 {
		fmt.Printf("  no tasks found\n")
	}

}

func subcommandInfo(ctx climax.Context) error {
	// parse args and things
	directory := DefaultDirectory
	if d, ok := ctx.Get("directory"); ok {
		directory = d
	}
	printProcessInformation(directory)
	printTaskInformation(directory)
	return nil
}
