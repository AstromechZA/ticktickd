package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/AstromechZA/ticktickrules"
)

// TaskDefinition defines the structure of the task deserialised from files
// on disk.
type TaskDefinition struct {
	// Name is a unique name of the task
	Name string `json:"name"`

	// Rule is a string containing the cron rule like "* * * * *"
	Rule string `json:"rule"`

	// Command is the array of strings to execute as a command
	Command []string `json:"command"`

	// RunAsUser defines the username (or uid) to execute the command as
	RunAsUser string `json:"runas"`
}

// GetRule extracts the cron rule from the task definition
func (td *TaskDefinition) GetRule() (*ticktickrules.Rule, error) {
	ruleParts := strings.Split(td.Rule, " ")
	if len(ruleParts) != 5 {
		return nil, fmt.Errorf("task Rule string must contain exactly 5 space-seperated parts")
	}
	return ticktickrules.NewRule(ruleParts[0], ruleParts[1], ruleParts[2], ruleParts[3], ruleParts[4])
}

// Validate validates the given task to check whether it is usuable
func (td *TaskDefinition) Validate() error {
	if len(td.Name) < 2 {
		return fmt.Errorf("task Name must be at least 2 characters long")
	}
	if len(td.Command) == 0 {
		return fmt.Errorf("task Command must have at least one element")
	}

	_, err := td.GetRule()
	if err != nil {
		return fmt.Errorf("task Rule is invalid: %s", err)
	}
	return nil
}

// LoadTaskDefinitions loads all of the task json files from the given directory.
// It returns the list of loaded definitions, a map of errors for failed task loads, and a main error if everything failed.
func LoadTaskDefinitions(tasksDirectory string) ([]TaskDefinition, map[string]error, error) {
	var definitions []TaskDefinition
	failures := make(map[string]error)
	fileinfos, err := ioutil.ReadDir(tasksDirectory)
	if err != nil {
		return definitions, failures, err
	}
	for _, fi := range fileinfos {
		if !fi.IsDir() {
			if strings.HasSuffix(fi.Name(), ".json") {
				taskBytes, ferr := ioutil.ReadFile(path.Join(tasksDirectory, fi.Name()))
				if ferr != nil {
					failures[fi.Name()] = fmt.Errorf("failed to read file: %s", ferr)
					continue
				}
				var td TaskDefinition
				if ferr := json.Unmarshal(taskBytes, &td); ferr != nil {
					failures[fi.Name()] = fmt.Errorf("failed to parse json: %s", ferr)
					continue
				}
				if ferr := td.Validate(); ferr != nil {
					failures[fi.Name()] = fmt.Errorf("failed validation: %s", ferr)
					continue
				}
				definitions = append(definitions, td)
			}
		}
	}
	return definitions, failures, nil
}
