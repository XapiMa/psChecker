# psChecker
This is a tool that provides process health checks.
Check if there are processes with the specified information.

## Installation
```
$ go get github.com/XapiMa/psChecker
```

or

```
$ git clone https://github.com/XapiMa/psChecker.git
$ go build main.go
```

If you need a different Architecture executable file:

```
 $ git clone https://github.com/XapiMa/psChecker.git
 $ GOOS=linux GOARCH=amd64 go build main.go -o psChecker
```
Please refer to the official document for details of available environment.
https://golang.org/doc/install/source#environment


## Usage
### show current processes
```
$ psChecker show

USER    PID     EXEC_FILE_NAME              ARGS 
root    4875    /sbin/auditd    
root    4988    /usr/sbin/NetworkManager    --no-daemon
root    5448    /usr/sbin/sshd              -D
```


### Write config.yml
Write the process information of health check target in config.yml

```
alive:
  - user: root
    pid: 4875
    exec: /sbin/auditd
    args: ""
  - exec: /usr/sbin/NetworkManager
    args: --no-daemon
  - pid: 5448
    exec: /usr/sbin/sshd
    args: -D
dead:
  - regexp: .*backdoor.*
  - regexp: .*crack.*
```

Warn when there is no process with the value set to alive and when there is a process with the value set to dead.

Possible values are user, pid, exec, args and regexp.
- user: Execution user name
- pid: Process ID
- exec: executable file path
- args: command line arguments at runtime
- regexp: regular expression to search for all user, id, exec, args


## Execution

```
$ psChecker check -t path/to/config.yml
```

If you want to write the result to a file:
```
$ psChecker -t path/to/config.yml -o path/to/output/file
```

```
Usage of psChecker:
        -n int
                Parallel number. (default 200)
        -o string
                output file path. If not set, it will be output to standard output
        -t string
                path to config.yml (default "In the same directory as the executable file")
```

