Keywords: Golang, go, concurrency, JSON

<img src="gopher.png" alt="Golang logo" align="right"/>

# {GpX} Go Concurrent Processes Executer
[![][license-svg]][license-url]

[*Concurrency is not paralelism*](https://blog.golang.org/concurrency-is-not-parallelism)

<hr>

## What is *gcpex* ?

Messing around `*nix` , a need arises frequently for me, **concurrent** execution of multiple, **non related** proccesses.

Each one must be launched with their *own parameters* and directed to their *own custom log* Files.

Umm ... Just use a  `bash` Script *you idiot* :neckbeard: ...

Well, that was my First approach and It worked *nicely* ...  but the code was *kind of ugly and cumbersome* .. not to mention It's relative poor performance.

During my **Go** learning journey I read this [**article**](https://blog.golang.org/pipelines) that Came to my Mind Naturally when trying to solve this task.

Based on that knowledge I built my own tool `gcpex`.

:point_right: **Note**: Only [`stdlib`](https://golang.org/pkg/#stdlib) packages and **just** one [`source`](https://github.com/klashxx/gcpex/blob/master/main.go) file used.

## Demo
[![demo][asciicast-png]][asciicast-url]

## Tell me about the installation

Obviously you need [`go`](https://golang.org/doc/install) in your machine.

Then just:

```bash
$ go get -v github.com/klashxx/gcpex
````

And the executable will be compiled and placed in your `$GOPATH/bin` directory.

```
$ which gcpex
/Users/klashxx/Documents/dev/go/bin/gcpex
```

Easy enough :sunglasses:

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

  - `cmd`: Executable {**mandatory**}
  - `args`: List of arguments to parse to the executable {optional}
  - `log`: Path to the log File attached to `cmd`  `stdout` and `stderr`. {optional} (missed if not specified)
  - `env`: List of environment variables to use for launch the process, if `env` is `null` it uses the current environment
  - `overwrite`: A `bool` value, must be switched to `true` to *overwrite* a previous log File. {optional} (default = `false`)

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

3. `-routines`: number of *routines* to *digester* the commands stored in our `JSON` `-in` File.

<hr>

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
2017/02/03 00:12:46 Start -> Cmd: echo          Args: 5               Pid:  8845
2017/02/03 00:12:46 Start -> Cmd: ls            Args: -j              Pid:  8846
2017/02/03 00:12:46 End   -> Cmd: echo          Args: 5               Pid:  8845 Success: true  Elapsed: 0000
2017/02/03 00:12:46 ERROR -> Cmd: ls            Args: -j              Err: exit status 1
2017/02/03 00:12:46 Start -> Cmd: sleep         Args: 5               Pid:  8847
2017/02/03 00:12:46 Start -> Cmd: sleep         Args: 5               Pid:  8848
2017/02/03 00:12:51 End   -> Cmd: sleep         Args: 5               Pid:  8848 Success: true  Elapsed: 0005
2017/02/03 00:12:51 End   -> Cmd: sleep         Args: 5               Pid:  8847 Success: true  Elapsed: 0005
2017/02/03 00:12:51 ERROR -> Cmd: dummy02       Args: 5               Err: exec: "dummy02": executable file not found in $PATH
2017/02/03 00:12:51 ERROR -> Cmd: cat           Args: commands.json   Err: /non_existent/commands.out: file base dir does not exists
2017/02/03 00:12:51 Start -> Cmd: echo          Args: Lorem ipsum dolor sit amet Pid:  8849
2017/02/03 00:12:51 End   -> Cmd: echo          Args: Lorem ipsum dolor sit amet Pid:  8849 Success: true  Elapsed: 0000
2017/02/03 00:12:51 Final -> Elapsed (seconds): 0005                  Executions (tot/ok/ko): 007 / 004 / 003
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

Content of the result file `response.json`:

```json
[
{
  "Cmd": "echo",
  "Path": "/bin/echo",
  "Env": null,
  "Args": [
    "5"
  ],
  "Success": true,
  "Pid": 8845,
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
  "Pid": 8846,
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
  "Pid": 8848,
  "Duration": 5,
  "Errors": null,
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
  "Pid": 8847,
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
    "Lorem ipsum dolor sit amet"
  ],
  "Success": true,
  "Pid": 8849,
  "Duration": 0,
  "Errors": null,
  "Log": "/tmp/commands.out",
  "Overwrite": false
}
]
```
<hr>

### Nice? Let's try Another one ....

Suppose a [`commands_02.json`](https://github.com/klashxx/gcpex/blob/master/samples/commands_02.json) file with **30** `sleep 5` *processes*:

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
$ gcpex -in commands_02.json -routines 10
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7961
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7960
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7962
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7966
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7967
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7963
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7968
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7969
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7965
2017/02/03 00:03:53 Start -> Cmd: sleep         Args: 5     Pid:  7964
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7960 Success: true  Elapsed: 0005
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7961 Success: true  Elapsed: 0004
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7970
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7971
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7964 Success: true  Elapsed: 0005
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7962 Success: true  Elapsed: 0005
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7963 Success: true  Elapsed: 0005
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7972
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7973
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7965 Success: true  Elapsed: 0005
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7966 Success: true  Elapsed: 0005
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7974
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7967 Success: true  Elapsed: 0005
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7975
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7976
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7977
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7968 Success: true  Elapsed: 0005
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7978
2017/02/03 00:03:58 End   -> Cmd: sleep         Args: 5     Pid:  7969 Success: true  Elapsed: 0005
2017/02/03 00:03:58 Start -> Cmd: sleep         Args: 5     Pid:  7979
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7970 Success: true  Elapsed: 0005
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7971 Success: true  Elapsed: 0005
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7980
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7981
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7972 Success: true  Elapsed: 0005
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7973 Success: true  Elapsed: 0005
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7974 Success: true  Elapsed: 0005
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7982
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7983
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7975 Success: true  Elapsed: 0005
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7976 Success: true  Elapsed: 0005
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7984
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7978 Success: true  Elapsed: 0005
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7977 Success: true  Elapsed: 0005
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7985
2017/02/03 00:04:03 End   -> Cmd: sleep         Args: 5     Pid:  7979 Success: true  Elapsed: 0005
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7986
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7987
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7988
2017/02/03 00:04:03 Start -> Cmd: sleep         Args: 5     Pid:  7989
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7980 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7981 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7982 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7983 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7986 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7985 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7984 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7987 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7988 Success: true  Elapsed: 0005
2017/02/03 00:04:08 End   -> Cmd: sleep         Args: 5     Pid:  7989 Success: true  Elapsed: 0005
2017/02/03 00:04:08 Final -> Elapsed (seconds): 0015        Executions (tot/ok/ko): 030 / 030 / 000
```

As expected the total execution time is 15 seconds.

Now ... We're going to use the :horse_racing: *Calvary*.

**Thirty routines** in action:

```
$ gcpex -in commands_02.json -routines 30
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8102
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8104
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8105
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8106
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8103
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8109
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8110
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8111
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8107
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8108
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8112
2017/02/03 00:05:39 Start -> Cmd: sleep         Args: 5     Pid:  8113
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8114
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8115
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8116
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8117
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8118
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8119
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8120
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8121
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8122
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8123
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8124
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8125
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8126
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8127
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8128
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8129
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8130
2017/02/03 00:05:40 Start -> Cmd: sleep         Args: 5     Pid:  8131
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8102 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8105 Success: true  Elapsed: 0004
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8103 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8104 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8106 Success: true  Elapsed: 0004
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8109 Success: true  Elapsed: 0004
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8107 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8108 Success: true  Elapsed: 0004
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8110 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8112 Success: true  Elapsed: 0005
2017/02/03 00:05:44 End   -> Cmd: sleep         Args: 5     Pid:  8111 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8113 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8114 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8115 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8116 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8117 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8118 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8119 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8120 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8121 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8122 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8123 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8124 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8125 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8126 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8128 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8127 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8129 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8130 Success: true  Elapsed: 0005
2017/02/03 00:05:45 End   -> Cmd: sleep         Args: 5     Pid:  8131 Success: true  Elapsed: 0005
2017/02/03 00:05:45 Final -> Elapsed (seconds): 0005        Executions (tot/ok/ko): 030 / 030 / 000
```

Again... the result **makes sense**, the program *took five seconds* to process it all.

## Licensing
**gcpex** is licensed under the MIT License. See [LICENSE](https://github.com/klashxx/gcpex/blob/master/LICENSE) for the full license text.

## Contact me

You can find me out [**here**](https://klashxx.github.io/about) :godmode:

<center><h6 align="center">
<br>Made with :heart: in <a href="https://www.google.com/search?q=almeria&espv=2&biw=1217&bih=585&sa=X#tbm=isch&q=almeria+movies">Almería</a>, Spain.
</h6></center>

[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://opensource.org/licenses/MIT

[asciicast-png]: https://asciinema.org/a/132235.png
[asciicast-url]: https://asciinema.org/a/132235
