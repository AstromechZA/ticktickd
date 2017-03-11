package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usageString = `
TODO
`

const logoImage = `
  __________  ____  ____
 /_  __/ __ \/ __ \/ __ \
  / / / / / / / / / / / /
 / / / /_/ / /_/ / /_/ /
/_/  \____/_____/\____/
`

// These variables are filled by the `govvv` tool at compile time.
// There are a few more granular variables available if necessary.
var Version = "<unofficial build>"
var GitSummary = "<changes unknown>"
var BuildDate = "<no date>"

func mainInner() error {

	// first set up config flag options
	versionFlag := flag.Bool("version", false, "Print the version string")

	// set a more verbose usage message.
	flag.Usage = func() {
		os.Stderr.WriteString(strings.TrimSpace(usageString) + "\n\n")
		flag.PrintDefaults()
	}
	// parse them
	flag.Parse()

	// first do arg checking
	if *versionFlag {
		fmt.Printf("Version: %s (%s) on %s \n", Version, GitSummary, BuildDate)
		fmt.Println(logoImage)
		return nil
	}

    fmt.Println("Hello World")

	return nil
}

func main() {
	if err := mainInner(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
