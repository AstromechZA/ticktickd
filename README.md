# `ticktickd` - cron-like task runner

`ticktickd` is a simple binary for running tasks on various intervals. It can be run separately by any
user in a working directory which is used to store task definitions, a last run time
database, and a rotating log file.

## Commands available

```
$ ./ticktickd
'ticktickd' is a simple binary for running tasks on various intervals. It can be run separately by any
user in a working directory which is used to store task definitions, a last run time
database, and a rotating log file.

Usage:

	ticktickd command [arguments]

The commands are:

	version     print version information
	run         run the ticktickd process (not daemonized)
	signal      signal the ticktickd process to reload its tasks
	info        print information about the task and run times

Use "ticktickd help [command]" for more information about a command.
```

The `info` command can be particularly useful as it gives you information like:

```
$ ./ticktickd info -d ~/.config/ticktickd
Process:
  ticktickd is not running (no pidfile)
Tasks:
  a.json:
    name: 1-minutely
    rule: * * * * *
    last run: 2017-03-21 17:41:14.445201323 +0200 SAST
    next run: 2017-03-21 19:30:00 +0200 SAST (in 48.399766008s)
  b.json:
    name: daily
    rule: * 12 * * *
    last run: never
    next run: 2017-03-22 12:00:00 +0200 SAST (in 16h30m48.399766008s)
```

Definitely not something `cron` can usually do!

## Installation

1. Download the binary appropriate for your system.

2. If you're installing it as root place it in `/usr/local/bin` or any other appropriate location. If
not, `~/bin/ticktickd` is probably a good place.

3. Create the working directory to run it out of. The default working directory is `/etc/ticktickd` which
will work if you're root or running it as an appropriate user. If not, `~/.config/ticktickd` is probably
best.

4. Add `ticktickd` as a reliable service via an init system like `systemd`, `supervisord`, `upstart` etc.
This will keep it running and start it at boot. The command should be `/usr/local/bin/ticktickd run` or
something like `~/bin/ticktickd run --directory ~/.config/ticktickd`.

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
