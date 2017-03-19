package main

import (
	"os/exec"
	"path"
	"sort"
	"time"
)

type taskNextTime struct {
	taskDefinition TaskDefinition
	nextRunTime    time.Time
}

func doWork(directory string) (sleeptime time.Duration) {
	// this is the current time to be used throughout this work iteration
	workTime := time.Now()

	// default sleep seconds to 5 minutes in case of unexpected errors
	sleeptime = time.Duration(5) * time.Minute

	db, err := InitTimeDB(directory)
	if err != nil {
		log.Errorf("critical error when opening database: %s", err)
		return
	}
	defer db.Close()
	EnsureBucket(db)

	tasksDir := path.Join(directory, "tasks.d")
	log.Infof("Loading tasks from %s..", tasksDir)
	tasks, loadfailures, err := LoadTaskDefinitions(tasksDir)
	if err != nil {
		log.Errorf("Critical failure when loading tasks: %s", err)
		return
	}
	for name, ferr := range loadfailures {
		log.Errorf("Error while loading task from file %s: %s", name, ferr)
	}
	log.Infof("Loaded %d tasks successfully.", len(tasks))

	var tasksToSpawn []TaskDefinition

	// now construct task time things
	var tasksToWaitFor []taskNextTime
	for _, td := range tasks {
		r, _ := td.GetRule()
		lastRunTime := GetLastRunTime(db, &td)
		if lastRunTime.IsZero() {
			// has never run before
			if r.Matches(workTime) {
				tasksToSpawn = append(tasksToSpawn, td)
			}
			nextTime := r.NextAfter(workTime)
			tasksToWaitFor = append(tasksToWaitFor, taskNextTime{td, nextTime})
		} else {
			// has run before
			nextAfterLast := r.NextAfter(lastRunTime)
			if nextAfterLast.Before(workTime) && r.Matches(workTime) {
				tasksToSpawn = append(tasksToSpawn, td)
			}
			nextTime := r.NextAfter(workTime)
			tasksToWaitFor = append(tasksToWaitFor, taskNextTime{td, nextTime})
		}
	}

	// spawn the matching tasks!
	for _, td := range tasksToSpawn {
		StoreLastRunTime(db, &td, workTime)
		exe := td.Command[0]
		args := td.Command[1:]
		cmd := exec.Command(exe, args...)
		log.Infof("Spawning %s..", td.Name)
		err := cmd.Start()
		if err != nil {
			log.Errorf("Failed to spawn process %s %s: %s", exe, args, err)
		}
	}

	if len(tasksToWaitFor) > 0 {
		sort.Slice(tasksToWaitFor, func(i, j int) bool { return tasksToWaitFor[i].nextRunTime.Before(tasksToWaitFor[j].nextRunTime) })
		nextTask := tasksToWaitFor[0]
		waitTime := nextTask.nextRunTime.Sub(workTime)
		log.Infof("Next task '%s' should run at %s (in %s)", nextTask.taskDefinition.Name, nextTask.nextRunTime, waitTime)
		sleeptime = sleepTimeFromWaitTime(waitTime)
	}
	log.Infof("Will sleep for %s", sleeptime)
	return
}

func sleepTimeFromWaitTime(waitTime time.Duration) time.Duration {
	if waitTime < time.Minute {
		return waitTime
	} else if waitTime < 5*time.Minute {
		return time.Minute
	} else if waitTime < 30*time.Minute {
		return 5 * time.Minute
	}
	return 30 * time.Minute
}
