package main

import (
	"log"
	"os/exec"
	"path"
	"sort"
	"syscall"
	"time"

	"github.com/AstromechZA/etcpwdparse"
)

type taskNextTime struct {
	taskDefinition TaskDefinition
	nextRunTime    time.Time
}

func doWork(directory string) (sleeptime time.Duration) {
	// this is the current time to be used throughout this work iteration
	workTime := time.Now()

	// default sleep seconds to 5 minutes in case of unexpected errors
	sleeptime = 5 * time.Minute

	db, err := InitTimeDB(directory)
	if err != nil {
		log.Printf("critical error when opening database: %s", err)
		return
	}
	defer db.Close()
	EnsureBucket(db)

	tasksDir := path.Join(directory, "tasks.d")
	tasks, loadfailures, err := LoadTaskDefinitions(tasksDir)
	if err != nil {
		log.Printf("Critical failure when loading tasks: %s", err)
		return
	}
	for name, ferr := range loadfailures {
		log.Printf("Error while loading task from file %s: %s", name, ferr)
	}
	log.Printf("Loaded %d tasks from %s.", len(tasks), tasksDir)

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

	cache, err := etcpwdparse.NewLoadedEtcPasswdCache()
	if err != nil {
		log.Printf("Failed to load etc password file, disabling any other-user tasks: %s", err)
		cache = nil
	}

	// spawn the matching tasks!
	for _, td := range tasksToSpawn {
		StoreLastRunTime(db, &td, workTime)
		exe := td.Command[0]
		args := td.Command[1:]
		cmd := exec.Command(exe, args...)

		if td.RunAsUser != "" {
			if cache == nil {
				log.Printf("Skipping task '%s' since we cant run as another user", td.Name)
				continue
			}
			entry, ok := cache.LookupUserByName(td.RunAsUser)
			if !ok {
				log.Printf("Skipping task '%s' since no user %s exists", td.Name, td.RunAsUser)
				continue
			}
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(entry.Uid()), Gid: uint32(entry.Gid())}
		}

		log.Printf("Spawning %s..", td.Name)
		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to spawn task %s process %s %s: %s", td.Name, exe, args, err)
		}
	}

	if len(tasksToWaitFor) > 0 {
		sort.Slice(tasksToWaitFor, func(i, j int) bool { return tasksToWaitFor[i].nextRunTime.Before(tasksToWaitFor[j].nextRunTime) })
		nextTask := tasksToWaitFor[0]
		waitTime := nextTask.nextRunTime.Sub(workTime)
		log.Printf("Next task '%s' should run at %s (in %s)", nextTask.taskDefinition.Name, nextTask.nextRunTime, waitTime)
		sleeptime = sleepTimeFromWaitTime(waitTime)
	} else {
		// otherwise we just sleep for 30 minutes
		sleeptime = 30 * time.Minute
	}
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
