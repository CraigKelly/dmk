# dmk

This is a simplified, automated build tool for data projects.

The idea behind dmk is to support build files that are easy to read *and*
write, and to support automating a build system for data artifacts. `dmk` was
inspired by `make`, `scons`, and `fac`.

For each command in a pipeline, you need to supply:

* A name
* The actual command to run (in the shell)
* The inputs required
* The outputs generated
* (*Optionally*) a list of intermediate files that `clean` process can delete

A file named `Pipeline` specifies these commands in YAML. All build steps run
in parallel, but each step waits until other steps build its dependencies. A
single build step executes the following steps:

1. The step is "Started"
2. If any of the required inputs are another step's outputs, then wait for a built message.
3. Check to see if *any* outputs are older than *any* of the inputs. If not, then the step is "Completed"!
4. If not done, set status to "Executing" and run the command.
5. If the command returns an error code or if *any* outputs are missing or older than *any* inputs, the step is "Failed".
6. If not done, send notification messages for each output for any waiting steps.
7. The step is now "Completed"

The outputs for a step must be unique to that step: you can't have two steps
both list `foo.data` as an output.


## Who should use this?

This is a tool for data flows and simple projects. Often these projects
involve one-time steps for getting the data used for analysis. Often the user
manually places that data in the directory after downloading it from Amazon S3
or a research server or whatever. Scripts or programs run in a pipeline
fashion: handling cleaning, transformation, analysis, model building,
figure production, presentation building, etc.

Pipelines like this are often written Python or R, partially automated with
shell scripts, and then tied together with a Makefile (or a SConstruct file if
you're an `scons` fan).

## What is this NOT for?

This is *not* mean to replace a real automated build tool for a software
project. As a general rule:

* If you're build Go software, use the Go tools (and optionally make)
* If you're building Java/Scala use `sbt`, `gradle`, `mvn`, `ant`, etc
* If you're building .NET, erm, I'm not sure
* There are great tools like `scons` that understand how to build lots of things (including LaTeX docs)
* If you're not sure, at least understand why you wouldn't just use `make`
