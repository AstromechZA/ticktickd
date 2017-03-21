package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/AstromechZA/ticktickd/pidfile"
	"github.com/tucnak/climax"
)

func subcommandSignal(ctx climax.Context) error {

	// parse args and things
	directory := DefaultDirectory
	if d, ok := ctx.Get("directory"); ok {
		directory = d
	}

	// setup pidfile
	pf, err := pidfile.NewPidfile(directory, "ticktickd.pid")
	if err != nil {
		return fmt.Errorf("pidfile error: %s", err)
	}

	// read it
	pid, err := pf.Read()
	if err != nil {
		return fmt.Errorf("failed to read pidfile: %s", err)
	}
	if pid == pidfile.MissingPidFile {
		return fmt.Errorf("ticktickd process is not running or pidfile does not exist")
	} else if pid < 2 {
		return fmt.Errorf("invalid pid for process: %d", pid)
	}

	// find and signal the process
	proc, _ := os.FindProcess(pid)
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		if err == syscall.ESRCH {
			return fmt.Errorf("ticktickd is not running")
		}
		return fmt.Errorf("unable to send signal to pid %d", pid)
	}
	if err := proc.Signal(syscall.SIGUSR1); err != nil {
		return fmt.Errorf("could not send SIGUSR1 to process %d", pid)
	}
	fmt.Printf("Signalled running ticktickd process %d\n", pid)
	return nil
}
