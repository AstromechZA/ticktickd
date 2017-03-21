# `ticktickd` - cron-like task runner

`ticktickd` is a simple binary for running tasks on various intervals. It can be run separately by any
user in a working directory which is used to store task definitions, a last run time
database, and a rotating log file.

## Installation

1. Download the binary appropriate for your system.

2. If you're installing it as root place it in `/usr/local/bin` or any other appropriate location. If
not, `~/bin/ticktickd` is probably a good place.

3. Create the working directory to run it out of. The default working directory is `/etc/ticktickd` which
will work if you're root or running it as an appropriate user. If not, `~/.config/ticktickd` is probably
best.

4. Add `ticktickd` as a reliable service via an init system like `systemd`, `supervisord`, `upstart` etc.
This will keep it running and start it at boot.

## Behavior under changing time

The thing that makes `cron` and similar software complicated is the fact that it must continue running tasks
predictably even when there are external influences to the system time like NTP or manual clock changes.

`ticktickd` takes a simpler best-effort approach on each wake-up:

1. the tasks are loaded, interpreted, and validated (invalid tasks are ignored)
2. if the cron rule matches the current time, the task is executed
3. the cron rule is used to calculate the _next_ execution time after the current time
4. the time until the soonest execution time (T) is used to sleep until the next wake-up:
    - if T < 1 minute, then sleep exact seconds until next time
    - if T < 5 minutes, then sleep 60 seconds
    - if T < 30 minutes, then sleep 5 minutes
    - otherwise, just sleep 5 minutes

The last run time is stored in a small database which prevents a task being executed twice in the event that the
clock is turned back.

Remember that the rule matches the entire window of execution, so a minutely rule will have 60 seconds to
be run in, while a hourly rule will have 60 minutes. So a minutely rule is robust to clock swings of up to 59
seconds while an hourly rule will still be run even if the clock is advanced by up to 59 minutes.

### Task Definitions

In the working directory, tasks are loaded from a `tasks.d/` directory. Any `.json` file is loaded
as a structure:

```
{
    "name": "2-minutely",
    "rule": "*/2 * * * *",
    "command": ["/some/command", "-a", "bob", "things"]
}
```

An optional `runas` key can be used to define a username to run the command under. The user running
the `ticktickd` daemon must be able to switch to the given user.

The rules are defined in the same was as cron:

- 5 parts (minute, hour, dayofmonth, monthofyear, dayofweek)
- '*' matches any value
- '*/N' matches 0 or a multiple or `N`
- 'N/M/..' matches `N`, `M`, etc..

We use [ticktickrules](https://godoc.org/github.com/AstromechZA/ticktickrules) to evaluate
the rules, so there is a bit more documentation under that page. More information about cron rules
can be found in the man pages for cron.

### Reloading Task Definitions

The task definitions are reloaded under 3 conditions:

1. sleep time expires (between 1 and 30 minutes depending on next soonest task execution)
2. SIGUSR1 signal received
3. inotify event from the `tasks.d/` directory

### Development:

This project uses a couple of useful dev dependencies:

```
$ go get github.com/ahmetalpbalkan/govvv
$ go get github.com/kardianos/govendor
```
