# {GpX} Go Concurrent Procceses Executer

[*Concurrency is not paralelism*](https://blog.golang.org/concurrency-is-not-parallelism)

<hr>

## What is *gcpex* ?

Messing around `*nix` , a need arises frequently for me, **concurrent** execution of multiple, **non related** proccesses.

Each one must be launched with their *own parameters* and directed to their *own custom log* Files.

Umm ... Just use a  `bash` Script *you idiot* :neckbeard: ...

Well, that was my First approach and It worked *nicely* ...  but the code was *kind of ugly and cumbersome* .. not to mention It's relative poor performance.

During my **Go** learning journey I read this [**article**](https://blog.golang.org/pipelines) that Came to my Mind Naturally when trying to solve this task.

Based on that knowledge I built my own tool `gcpex`.

## OK, now ... how does this thing works?

The syntax is neat:

```bash
$ gcpex
  -in string
    	cmd JSON file repo. [mandatory]
  -out string
    	Respond JSON file.
  -routines int
    	max parallel execution routines (default 5)
 ```

<hr>

1. `-in`: **Mandatory requisite**, a `JSON` File to Configure our *bunch of executions*.

The format is pretty self explanatory:

```json
[
 {
  "cmd": "a_command",
  "args": ["arg1", "arg2"],
  "log": "/my/log/path/a_command.log",
  "overwrite": true
 },
 {
  "cmd": "another_command",
  "env": ["PATH=/my/custom/path"],
  "log": "/non_existent/commands.out"
 }
 ]
```

Schema definition:

- `cmd`: Executable [**mandatory**]
- `args`: List of arguments to parse to the executable [optional]
- `log`: Path to the log File attached to `cmd`  `stdout` and `stderr`. [optional] (missed if not specified)
- `env`: List of environment variables to use for launch the process, if `env` is `null` it uses the current environment
- `overwrite`: A `bool` value, must be switched to `true` to *overwrite* a previous log File. [optional] (default = `false`)

<hr>

2. `-out`: an optional `JSON` file where the Response will be Written.

 Format:

```json
[
 {
  "Cmd": "a_command",
  "Path": "/path/to/command",
  "Env": null,
  "Args": [
    "arg1",
    "arg2"
  ],
  "Success": true,
  "Pid": 11111,
  "Duration": 15,
  "Errors": [],
  "Log": "/my/log/path/a_command.log",
  "Overwrite": true
},
{
  "Cmd": "another_command",
  "Path": "/my/custom/path",
  "Env": [
    "PATH=/my/custom/path"
  ],
  "Args": [],
  "Success": false,
  "Pid": 0,
  "Duration": 0,
  "Errors": [
    "/non_existent/commands.out: file base dir does not exists"
  ],
  "Log": "/non_existent/commands.out",
  "Overwrite": false
}
]
```

Schema definition:

- `Cmd`: Full path to the cmd executed
- `Path`: Dir path to executable.
- `Env`: List of environment variables used to launch the process.
- `Args`: List of arguments parsed to the executable.
- `Success`: A `bool` value, will be `true` when `cmd` exit code is 0.
- `Pid`: [*Process Identification Number*](http://www.linfo.org/pid.html) during the execution. Zero when process fails.
- `Duration`: Number of seconds exec took to complete.
- `Errors`: List of errors presented during the execution.
- `Log`: File used to store `stdout` and `stderr`.
- `Overwrite`: À `bool` flag, when `true` allowed to overwrite a previous Log file.

<hr>

3. `-routines`: number of *routines* to *digester* the commands stored in our `JSON` `-in` File.

## Examples

Having this [`commands_01.json`](https://github.com/klashxx/gcpex/blob/master/samples/commands_01.json) file:

```json
[
  {
  "cmd": "echo",
  "args": ["5"]
 },
 {
  "cmd": "ls",
  "args": ["-w"],
  "log": "/tmp/ls.out"
 },
 {
  "cmd": "sleep",
  "args": ["5"]
 },
 {
  "cmd": "sleep",
  "args": ["5"]
 },
  {
  "cmd": "dummy02",
  "args": ["5"]
 },
  {
  "cmd": "cat",
  "args": ["commands.json"],
  "log": "/tmp/commands.out"
 },
  {
  "cmd": "cat",
  "args": ["commands.json"],
  "log": "/non_existent/commands.out"
 }
]
```

Using two routines to *digester* and storing the result in `reponse.json`:

```bash
2017/01/31 20:59:46 Start -> Cmd: echo Args: [5] PID: 54390
2017/01/31 20:59:46 End   -> Cmd: echo Args: [5] PID: 54390 Success: true
2017/01/31 20:59:46 Start -> Cmd: ls Args: [-j] PID: 54391
2017/01/31 20:59:46 Start -> Cmd: sleep Args: [5] PID: 54392
2017/01/31 20:59:46 ERROR -> Cmd: ls Args: [-j] Errors: [exit status 1]
2017/01/31 20:59:46 Start -> Cmd: sleep Args: [5] PID: 54393
2017/01/31 20:59:51 End   -> Cmd: sleep Args: [5] PID: 54392 Success: true
2017/01/31 20:59:51 ERROR -> Cmd: dummy02 Args: [5] Errors: [exec: "dummy02": executable file not found in $PATH]
2017/01/31 20:59:51 Start -> Cmd: echo Args: [Lorem ipsum dolor sit amet ,consectetur adipiscing elit,  sed do eiusmod tempor incididunt ut labore et dolore magna aliqua] PID: 54410
2017/01/31 20:59:51 End   -> Cmd: sleep Args: [5] PID: 54393 Success: true
2017/01/31 20:59:51 ERROR -> Cmd: cat Args: [commands.json] Errors: [/non_existent/commands.out: file base dir does not exists]
2017/01/31 20:59:51 End   -> Cmd: echo Args: [Lorem ipsum dolor sit amet ,consectetur adipiscing elit,  sed do eiusmod tempor incididunt ut labore et dolore magna aliqua] PID: 54410 Success: true
2017/01/31 20:59:51 Final -> Executions: 7 Success: 4 Fail: 3
2017/01/31 20:59:51 errors in execution/s
$ echo $?
1
```

Log of `ls` command:

```
$ cat /tmp/ls.out
STDOUT:
=======

<nil>

STDERR:
=======

ls: illegal option -- j
usage: ls [-ABCFGHLOPRSTUWabcdefghiklmnopqrstuwx1] [file ...]
```

Log of `echo` excution:

```
$ cat /tmp/commands.out
STDOUT:
=======

Lorem ipsum dolor sit amet ,consectetur adipiscing elit,  sed do eiusmod tempor incididunt ut labore et dolore magna aliqua

STDERR:
=======

<nil>
```

Content of the result file `respond.json`:

```json
cat response.json
[
{
  "Cmd": "echo",
  "Path": "/bin/echo",
  "Env": null,
  "Args": [
    "5"
  ],
  "Success": true,
  "Pid": 54390,
  "Duration": 0,
  "Errors": null,
  "Log": "",
  "Overwrite": false
},
{
  "Cmd": "ls",
  "Path": "/bin/ls",
  "Env": null,
  "Args": [
    "-j"
  ],
  "Success": false,
  "Pid": 54391,
  "Duration": 0,
  "Errors": [
    "exit status 1"
  ],
  "Log": "/tmp/ls.out",
  "Overwrite": false
},
{
  "Cmd": "sleep",
  "Path": "/bin/sleep",
  "Env": null,
  "Args": [
    "5"
  ],
  "Success": true,
  "Pid": 54392,
  "Duration": 5,
  "Errors": null,
  "Log": "",
  "Overwrite": false
},
{
  "Cmd": "dummy02",
  "Path": "",
  "Env": null,
  "Args": [
    "5"
  ],
  "Success": false,
  "Pid": 0,
  "Duration": 0,
  "Errors": [
    "exec: \"dummy02\": executable file not found in $PATH"
  ],
  "Log": "",
  "Overwrite": false
},
{
  "Cmd": "sleep",
  "Path": "/bin/sleep",
  "Env": null,
  "Args": [
    "5"
  ],
  "Success": true,
  "Pid": 54393,
  "Duration": 5,
  "Errors": null,
  "Log": "",
  "Overwrite": false
},
{
  "Cmd": "cat",
  "Path": "/bin/cat",
  "Env": null,
  "Args": [
    "commands.json"
  ],
  "Success": false,
  "Pid": 0,
  "Duration": 0,
  "Errors": [
    "/non_existent/commands.out: file base dir does not exists"
  ],
  "Log": "/non_existent/commands.out",
  "Overwrite": false
},
{
  "Cmd": "echo",
  "Path": "/bin/echo",
  "Env": null,
  "Args": [
    "Lorem ipsum dolor sit amet",
    ",consectetur adipiscing elit, ",
    "sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"
  ],
  "Success": true,
  "Pid": 54410,
  "Duration": 0,
  "Errors": null,
  "Log": "/tmp/commands.out",
  "Overwrite": false
}
]
```
