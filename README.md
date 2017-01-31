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
