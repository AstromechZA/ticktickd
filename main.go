package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tucnak/climax"
)

const descriptionString = `
'ticktickd' is a simple binary for running tasks on various intervals. It can be run separately by any
user in a working directory which is used to store task definitions, a last run time
database, and a rotating log file.
`

const logoImage = `
       .-.-.
  ((  (__I__)  ))
    .'_....._'.
   / / .12 . \ \
  | | '  |  ' | |
  | | 9  /  3 | |
   \ \ '.6.' / /
    '.'-...-'.'
     /'-- --'\
`

// Version string filled in by govvv
var Version = "<unofficial build>"

// GitSummary string filled in by govvv
var GitSummary = "<changes unknown>"

// BuildDate string filled in by govvv
var BuildDate = "<no date>"

// DefaultDirectory is the default directory
const DefaultDirectory = "/etc/ticktickd"

func main() {
	cli := climax.New("ticktickd")
	cli.Brief = strings.TrimSpace(descriptionString) + "\n\n\n"

	// Specify and add version command
	versionCmd := climax.Command{
		Name:  "version",
		Brief: "print version information",
		Handle: func(ctx climax.Context) int {
			fmt.Printf("Version: %s (%s) on %s \n", Version, GitSummary, BuildDate)
			fmt.Println(logoImage)
			fmt.Println("Project Url: https://github.com/AstromechZA/ticktickd")
			return 0
		},
	}
	cli.AddCommand(versionCmd)

	// Specify and add the run command
	runCmd := climax.Command{
		Name:  "run",
		Brief: "run the ticktickd process (not daemonized)",
		Flags: []climax.Flag{
			{
				Name:     "directory",
				Short:    "d",
				Usage:    "--directory=/some/dir",
				Help:     "set the working directory to load tasks, and write logs and pid files",
				Variable: true,
			},
			{
				Name:     "disablewatch",
				Usage:    "--disablewatch",
				Help:     "disable the inotify watch on the tasks directory, may improve performance",
				Variable: false,
			},
			{
				Name:     "nologfile",
				Usage:    "--nologfile",
				Help:     "don't log to the logfile, log everything to stdout",
				Variable: false,
			},
		},
		Handle: func(ctx climax.Context) int {
			if err := subcommandRun(ctx); err != nil {
				log.Printf("Error occured: %s", err)
				cli.Log(err.Error())
				return 1
			}
			return 0
		},
	}
	cli.AddCommand(runCmd)

	// Specify signal command
	signalCmd := climax.Command{
		Name:  "signal",
		Brief: "signal the ticktickd process to reload its tasks",
		Flags: []climax.Flag{
			{
				Name:     "directory",
				Short:    "d",
				Usage:    "--directory=/some/dir",
				Help:     "set the working directory to load tasks, and write logs and pid files",
				Variable: true,
			},
		},
		Handle: func(ctx climax.Context) int {
			if err := subcommandSignal(ctx); err != nil {
				fmt.Printf("An error occurred: %s\n", err)
				return 1
			}
			return 0
		},
	}
	cli.AddCommand(signalCmd)

	// Specify info command
	infoCmd := climax.Command{
		Name:  "info",
		Brief: "print information about the task and run times",
		Flags: []climax.Flag{
			{
				Name:     "directory",
				Short:    "d",
				Usage:    "--directory=/some/dir",
				Help:     "set the working directory to load tasks, and write logs and pid files",
				Variable: true,
			},
		},
		Handle: func(ctx climax.Context) int {
			if err := subcommandInfo(ctx); err != nil {
				fmt.Printf("An error occurred: %s\n", err)
				return 1
			}
			return 0
		},
	}
	cli.AddCommand(infoCmd)

	code := cli.Run()
	os.Exit(code)
}
