package main

import (
	"fmt"
	"os"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/tucnak/climax"
)

const descriptionString = `
TODO
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

var log = logging.MustGetLogger("ticktickd")

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
				Help:     "disable the inotify watch on the tasks directory",
				Variable: false,
			},
		},
		Handle: func(ctx climax.Context) int {
			directory := DefaultDirectory
			if d, ok := ctx.Get("directory"); ok {
				directory = d
			}

			if err := subcommandRun(directory, !ctx.Is("disablewatch")); err != nil {
				log.Criticalf("Error occured: %s", err)
				cli.Log(err.Error())
				return 1
			}

			return 0
		},
	}
	cli.AddCommand(runCmd)

	// Specify and add the signal command
	sigCmd := climax.Command{
		Name:  "signal",
		Brief: "signal the running ticktickd process to reload its tasks",
		Handle: func(ctx climax.Context) int {
			return 0
		},
	}
	cli.AddCommand(sigCmd)

	code := cli.Run()
	os.Exit(code)
}
