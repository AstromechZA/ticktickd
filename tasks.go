package main

// TaskDefinition defines the structure of the task deserialised from files
// on disk.
type TaskDefinition struct {
	Name      string   `json:"name"`
	Rule      string   `json:"rule"`
	Command   []string `json:"command"`
	RunAsUser string   `json:"runas"`
}
