package pidfile

import (
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/facebookgo/atomicfile"

	"strings"

	"io/ioutil"

	"strconv"

	"golang.org/x/sys/unix"
)

type Pidfile struct {
	directory string
	name      string
}

// NewPidfile constructs and validates a Pidfile object without creating a pidfile
// or checking for its existance.
func NewPidfile(directory string, name string) (*Pidfile, error) {

	// first a variety of checks whether the directory exists and is useful
	if st, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Directory path '%s' does not exist", directory)
		}
		return nil, fmt.Errorf("Failed to read directory '%s': %s", directory, err.Error())
	} else if !st.IsDir() {
		return nil, fmt.Errorf("Directory path '%s' is not a directory", directory)
	} else if unix.Access(directory, unix.W_OK) != nil {
		return nil, fmt.Errorf("Directory path '%s' is not writeable", directory)
	}

	// basic checks for the pidfile name
	if !strings.HasSuffix(name, ".pid") {
		return nil, fmt.Errorf("Pidfile name should end with .pid")
	}

	return &Pidfile{
		directory: directory,
		name:      name,
	}, nil
}

// Path returns the full pidfile path
func (p *Pidfile) Path() string {
	return path.Join(p.directory, p.name)
}

// MissingPidFile value returned by the Read  when the pidfile is missing
const MissingPidFile = -2 << 31

// Read the pidfile and pull out the value or error
// If the pidfile is missing it will return `MissingPidFile`
func (p *Pidfile) Read() (int, error) {
	f, err := os.Open(p.Path())
	if err != nil {
		if os.IsNotExist(err) {
			return MissingPidFile, nil
		}
		return 0, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return MissingPidFile, err
	}
	return v, nil
}

func (p *Pidfile) Write() error {

	// read the pidfile
	pid, err := p.Read()
	if err != nil {
		return fmt.Errorf("error reading pidfile: %s", err.Error())
	}

	// if the pidfile was not missing and we got a useful value
	if pid != MissingPidFile && pid > 1 {
		if p.isRunning(pid) {
			return fmt.Errorf("guarded process with pid %d is already running", pid)
		}
	}

	// open a pidfile
	file, err := atomicfile.New(p.Path(), os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("error opening pidfile %s: %s", p.Path(), err)
	}
	defer file.Close()

	// write the pid
	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	if err != nil {
		return err
	}

	return file.Close()
}

func (p *Pidfile) isRunning(pid int) bool {
	if proc, err := os.FindProcess(pid); err != nil {
		return false
	} else if err = proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}
