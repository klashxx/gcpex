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

## OK,now ... how does this thing works?

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

