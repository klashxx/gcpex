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
$ gcpex -in commands_01.json -routines 2 -out response.json
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
<hr>

### Nice? Let's try Another one ....

Suppose a [`commands_02.json`](https://github.com/klashxx/gcpex/blob/master/samples/commands_02.json) file with **30** `sleep 5` *proccesses*:

```json
[
 {
  "cmd": "sleep",
  "args": ["5"]
 },
 {
  "cmd": "sleep",
  "args": ["5"]
 },
 ...
 ]
 ```

Add **so on** ....

:checkered_flag: **Fact**: A sequential process would take **150 seconds** to complete.


Let's to use Ten **simultaneous** routines to do our Job:

```
$ time gcpex -in commands_02.json -routines 10
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55178
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55179
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55180
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55181
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55182
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55183
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55184
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55185
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55186
2017/01/31 21:10:44 Start -> Cmd: sleep Args: [5] PID: 55187
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55179 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55180 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55178 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55181 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55182 Success: true
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55191
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55192
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55193
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55194
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55184 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55183 Success: true
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55195
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55196
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55185 Success: true
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55197
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55198
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55186 Success: true
2017/01/31 21:10:49 End   -> Cmd: sleep Args: [5] PID: 55187 Success: true
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55199
2017/01/31 21:10:49 Start -> Cmd: sleep Args: [5] PID: 55200
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55192 Success: true
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55191 Success: true
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55193 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55212
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55194 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55213
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55195 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55214
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55215
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55196 Success: true
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55198 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55216
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55197 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55217
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55218
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55199 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55219
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55220
2017/01/31 21:10:54 End   -> Cmd: sleep Args: [5] PID: 55200 Success: true
2017/01/31 21:10:54 Start -> Cmd: sleep Args: [5] PID: 55221
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55213 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55212 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55214 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55215 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55216 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55217 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55218 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55219 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55220 Success: true
2017/01/31 21:10:59 End   -> Cmd: sleep Args: [5] PID: 55221 Success: true
2017/01/31 21:10:59 Final -> Executions: 30 Success: 30 Fail: 0

real    0m15.023s
user    0m0.008s
sys     0m0.027s
```

As expected the total execution time is 15 seconds.

:warning: **Note**: [`time`](https://linux.die.net/man/1/time) command is used to get *timing statistics*.

Now ... We're gong to use the :horse_racing: *Calvary*.

**Thirty routines** in action:

```
$ time gcpex -in commands_02.json -routines 10
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55798
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55799
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55801
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55797
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55800
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55802
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55803
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55804
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55805
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55806
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55807
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55808
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55809
2017/01/31 21:19:17 Start -> Cmd: sleep Args: [5] PID: 55810
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55811
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55812
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55813
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55815
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55814
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55816
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55817
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55818
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55819
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55820
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55821
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55822
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55823
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55824
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55825
2017/01/31 21:19:18 Start -> Cmd: sleep Args: [5] PID: 55826
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55797 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55799 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55798 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55801 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55800 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55802 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55803 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55804 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55805 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55806 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55807 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55808 Success: true
2017/01/31 21:19:22 End   -> Cmd: sleep Args: [5] PID: 55809 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55810 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55811 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55812 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55813 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55814 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55815 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55816 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55817 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55818 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55819 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55820 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55821 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55822 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55823 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55824 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55825 Success: true
2017/01/31 21:19:23 End   -> Cmd: sleep Args: [5] PID: 55826 Success: true
2017/01/31 21:19:23 Final -> Executions: 30 Success: 30 Fail: 0

real    0m5.030s
user    0m0.012s
sys     0m0.026s
```

Again .. the result **makes sense**, the program *took five seconds* to process it all.
