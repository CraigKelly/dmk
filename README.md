---
title: DMK ReadMe
---

![Go](https://github.com/CraigKelly/dmk/workflows/Go/badge.svg)

This is a simplified, automated build tool for data projects.

The idea behind dmk is to support build files that are easy to read *and*
write, and to support automating a build system for data artifacts. `make`,
`scons`, and `fac` all provided inspiration for `dmk`.

# Project Status

`dmk` does everything that I need it to do, and it is still my go-to project
for testing out CI/CD ideas. I'm planning on keeping it around for quite a
while because it's fun and it interests me. However, if you are here
because you need a tool like this, I think you'll be better served by taking
a look at [tasks](https://github.com/go-task/task):

* It's an active project with multiple contributors and sponsors
* It now has pretty much all the features I wanted for `dmk`
* It has some that I think are great (like an internal shell interpreter for Windows)

So go give them some love :)

# Who should use this?

This is a tool for data flows and simple projects. Often these projects
involve one-time steps for getting the data used for analysis. Often the user
manually places that data in the directory after downloading it from Amazon S3
or a research server or whatever. Scripts or programs run in a pipeline
fashion: handling cleaning, transformation, analysis, model building,
figure production, presentation building, etc.

Pipelines like this are often written Python or R, partially automated with
shell scripts, and then tied together with a Makefile (or a SConstruct file if
you're an `scons` fan).

*Protip*: if you're looking for a command to handle building reports from `.tex`
files (including handling metapost and biblatex), look into `rubber`.

# What is this NOT for?

This is *not* mean to replace a real automated build tool for a software
project. As a general rule:

* If you're building Go software, use the Go tools (and optionally make)
* If you're building Java/Scala use `sbt`, `gradle`, `mvn`, `ant`, etc
* If you're building .NET, erm, I'm not sure
* There are great tools like `scons` that understand how to build lots of artifacts (including LaTeX docs)
* If you're not sure, at least understand why you wouldn't use `make`

For instance, this project (written in Go) is actually built with `make` + the
standard Go tools.

# Using

When running `dmk`, you may specify `-h` on the command line for information
on command line parameters. Specify `-v` for "verbose" mode to get output you
may want for debugging or understanding what is going on.

For each command in a pipeline, you need to supply:

* A name
* The actual command to run (in the shell)
* The inputs required
* The outputs generated

This list is not exhaustive; see below for everything you can specify for a
build step.

The file is generally named `Pipeline` or `pipeline.yaml`. If you do not
specify a pipeline file with the `-f` command line parameter, `dmk` looks for
the following names in the current directory (in order):

* Pipeline
* Pipeline.yaml
* pipeline
* pipeline.yaml
* .Pipeline.yaml
* .pipeline.yaml

You may also supply a custom name with the `-f` command line flag. If the
pipeline file is in a different directory, dmk will change to that directory
before parsing the config file. Note that the bash tab completion logic will
show you step names, but it isn't smart enough to know that you've specified
`-f` on the command line.

All build steps run in parallel, but each step waits until other steps build
its dependencies. A single build step executes the following in order:

1. The step is "Started"
2. If any of the required inputs are another step's outputs, then wait for a built message.
3. Check to see if *any* outputs are older than *any* of the inputs. If not, then the step is "Completed"!
4. If not done, set status to "Executing" and run the command.
5. If the command returns an error code or if *any* outputs are missing or older than *any* inputs, the step is "Failed".
6. Send notification messages for each output for any waiting steps.
7. The step is now "Completed"

The outputs for a step must be unique to that step: you can't have two steps
both list `foo.data` as an output. (Note that this applies to *expanded* output
names, and abstract/baseSteps aren't checked.)

`dmk` provides an automatic "clean" mode that deletes all outputs. To use it,
specify `-c` on the command line. `dmk` will delete all the outputs for all
steps. If you have files to clean not specified as outputs, you can specified
them in the _clean_ list for a build step (see the Pipeline file format
below). Good candidates for the _clean_ section are intermediate files (such
as logs) generated as part of a build process that are not dependencies and
should not determine if a build step is up to date.

You may also run `dmk` with `-listSteps` to see a list of all steps in the current
pipeline file. Currently, this is used for bash completion.

# Pipeline file format

The file is in YAML format where each build step is a named hash. Each build
step should specify:

* _command_ - The command to run as part of the build. `dmk` uses bash to run
  the command, so it can rely on bash shell niceties (like using `~` for the
  home directory)
* _inputs_ - a list of inputs needed for the build. These are also the
  dependencies that must exist before the step can run. An entry can be a
  glob pattern (like `*.txt`)
* _outputs_ - a list of outputs generated by the step. Outputs decide if the
  step must run, and the clean phase deletes them. Glob patterns are
  **ignored** for outputs.
* _clean_ - A list of files to clean. These and outputs are the files deleted
  during a clean. You may use glob patterns for these.
* _explicit_ - Optional, defaults to false. If set to true, the step will
  run if you specify it on the command line. It will not run by default. Any steps
  required by steps specified on the command line will also run, regardless of their
  _explicit_ setting.
* _delOnFail_ - Optional, defaults to false. If set to true and the step fails,
  then `dmk` will delete all the step's output files.
* _direct_ - Optional, default to false. If set to true, both stdout and stderr
  from the step are written to the `dmk` process standard streams. If set to false
  (the default), stdout and stderr are written in single blocks after the step
  completes (stdout is only written if `dmk` is running in *verbose* mode).
  Note in *direct* mode (direct=True), step output may be interleaved with
  "normal" output when steps are running in parallel!
* _abstract_ - Optional, defaults to false. If specified, the step is an "base step"
  and will never be executed (it's only to be used as a baseStep). See below.
* _baseStep_ - Optional, defaults to empty. If specified, it must be the name of a
  step with `abstract: true`. In that case the step's properties will be based
  on the step given. See below.
* _vars_ - Optional, defaults to empty dictionary. If specified, this must be a
  hash/dictionary with strings as both keys and values. The keys are treated as
  variables names with are replaced with their corresponding values. See below
  for variable details.

The `res` subdirectory contains sample Pipeline files (used for testing), but
a quick example would look like:

````yaml
# You can have comments in a file
step1:                                # first step
    command: "xformxyz i{1,2,3}.txt"  # command with some shell magic
    inputs:                           # 3 inputs (read by our imaginary command)
        - i1.txt
        - i2.txt
        - i3.txt
    outputs:                          # 3 outputs
        - o1.txt
        - o2.txt
        - o3.txt
    clean: [a.aux, b.log]             # two extra clean targets, specified in
                                      # an alternate syntax for YAML lists

step2:                                # second step
    command: cmd1xyz                  # note the lack of inputs - this means
    outputs:                          # the step will run without waiting for
        - output.bin                  # other steps to complete.

depstep:                              # third/final step: it won't run until the
    command: cmd2xyz                  # previous steps finish because their
    inputs:                           # outputs are in the this step's inputs.
        - o3.txt
        - output.bin
    outputs:
        - combination.output
    clean:
        - need-cleaning.*             # An example of using a glob pattern
    delOnFail: true

extrastep:
    command: special-command
    inputs:
        - some-script-file.txt
    outputs:
        - my-special-file.extra
    explicit: true                    # Run if specified on command line (and not by default)
````

If you were to run `dmk -c` then it would deleted the following files:

* o1.txt, o2.txt, o3.txt because they are outputs of `step1`
* a.aux and b.log because they are in the `clean` list in `step1`
* output.bin because of `step2`
* combination.output and any files matching the pattern `need-cleaning.*` because of `depstep`

Note that my-special-file.extra from `extrastep` is not deleted unless you specify
`extrastep` on the command line.

After cleaning, if you run `dmk` the following steps would occur:

* The commands from `step1` (`xformxyz i{1,2,3}.txt`) and `step2` (`cmd1xyz`)
  would run
* When they were both finished, `depstep` would start and `cmd2xyz` would run.
* As before, `extrastep` would NOT run.
* If the `depstep` command (`cmd2xyz`) fails, then `dmk` will delete
  `combination.output` (if it exists).
* If all the steps succeed, running `dmk` again would not cause
  any command to run (because all outputs are newer than their steps' inputs).

If you were to run `dmk extrastep` then the command `special-command` would run.
Nothing else would run.

If you were to run `dmk extrastep depstep` then all steps would run (because
`step1` and `step2` are `depstep` dependencies).

# Using Variables

`dmk` steps support variable expansion.

Before variable expansion begins, both `inputs` and `clean` are expanded via
globbing (e.g. `*.csv` expands to all files ending in `.csv` in the current
directory.)

After globbing expansion, `dmk` will expand variables for all the strings in:

* command
* inputs
* outputs
* clean

Variables are expanded in the following order:

1. Any keys from the current step's `vars` section (if specified)
2. Any from the `vars` section of the current step's `baseStep` *if* that
   variable wasn't specified by the current step.
3. Any environment variables - note that you can set environment variables
   from an env file: see "Build Step Environment" below.

*IMPORTANT*: `DMK_STEPNAME` is defined at this point, but the other `DMK_`
variables described below in "Build Step Environment" are *not*. However,
the command will be executed in bash and they can be evaluated/used by a
script at run time.

See "Abstract/Base Steps" below for an example.

(See below for more explanation of base steps)

# Abstract/Base Steps

`dmk` provides a way to create small template steps so that you can simplify
pipeline files.  Often you'll have a few steps that have similar structures. In
that case you can specify a step with `abstract: true`. These steps will never
be executed, but provide a "template" for "concrete" steps.

If a step specified another step with `baseStep` then:

* If the step has command specified, it takes the command of its base step
* The base step's values for `explicit`, `delonFail`, and `direct` are all
  used, regardless of the child step's settings
* The base step's `inputs`, `outputs`, and `clean` entries are all added to
  the child step's lists.
* The base step's `vars` section provides the "defaults" for the child step
  (The child's `vars` section always wins)

Some rules:

* A step named in `baseStep` must have `abstract: true`
* An abstract step may *not* specify a `baseStep`


Example (note that `inputs` and `outputs` are missing):

```yaml
base:
    command: "echo $A $B $C"
    abstract: true
    vars:
        - A: Hello
        - B: World
stepa:
    baseStep: base
    vars:
        - B: There
        - C: Everyone
stepb:
    command: "echo $A $B $C"
    vars:
        - B: Anything
        - C: Missing
```

When `stepa` is executed, it will echo `Hello There Everyone` because it inherits
`A` from it's `baseStep`. Note that it does *not* use `B` from its `baseStep`.

When `stepb` is executed it will echo the string `Anything Missing` because the
command `echo $A Anything Missing` will be executed by bash, which will expand
`$A` to an empty string.

# Build Step Environment

Before reading the pipeline file, `dmk` will load the env file specified by the
(optional) `-e` command line parameter. This functionality comes from the
excellent [GoDotEnv](https://github.com/joho/godotenv) library, which is based
on the Ruby dotenv project.

When a build step runs, `dmk` sets environment variables in the step command's
process:

* DMK_VERSION - version string for dmk
* DMK_PIPELINE - absolute path to the pipeline file running
* DMK_STEPNAME - the name of the current step
* DMK_INPUTS - a colon (":") delimited list of inputs for this step
* DMK_OUTPUTS - a colon (":") delimited list of outputs for this step
* DMK_CLEAN - a colon (":") delimited list of extra clean files for this step

**IMPORTANT!** These `DMK_` variables are setup *after* config file processing
and *will* override any variables set in the environment before startup or via
an env file.

Also note that although `bash` evaluates the command, `dmk` does it's own
variable expansion before executing the command. However, only `DMK_STEPNAME`
will be defined for `dmk` variable expansion. See "Using Variables" above for
details.

When the command for a step is activated, it will inherit the original
environment that `dmk` is running in, modified in this order:

1. Start with the original environment
2. Set DMK_VERSION
3. Optionally load an .env file, which will update the environment
4. Set DMK_PIPELINE
5. For each step, add the build step environment variables
6. For each step, add the step variables (as defined above)


# Some helpful hints to remember

A pipeline file is a YAML document, and a **JSON** document is valid YAML. For
instance, `res/slowbuild.yaml` and `res/slowbuild.json` are semantically
identical pipeline files. If you need a customized build, you can generate the
pipeline file in the language of your choice in JSON or YAML and then call
`dmk`. As example, if you have a script named custom.py that outputs a JSON
pipeline on stdout, you can run the JIT pipeline with: `python3 custom.py | dmk -f -`.

Commands run in a new bash shell (which also means you need bash).

`dmk` changes to the directory of the Pipeline file, so you can specify file
names relative to the Pipeline file's directory. Of course, the current directory
is not changed if the Pipeline file is *stdin*.

You may use globbing patterns for the inputs and clean.

# Building

Currently, dependencies are managed with go modules. Before that dependencies
were vendored and managed with `dep`.  And before *that*, dependencies were
managed with `godep`. Both were excellent tools, but now we have a standard way
to manage dependencies in Go - and we no longer have a vendor dir.

You shouldn't need to worry about dependencies if you are building with the
`Makefile`. Also note the fact that we use `make` to build `dmk`. We are serious
about using the correct build tool for the job.

Yes, we currently regenerate `version.go` too frequently. If we ever get a single
contributor or pull request, we'll make it better :)

You should also have Python 3 installed (for `script/versiongen` and for the test
script `res/slow`).

`make dist` will build cross-platform binaries in `./dist`. Yes, we commit them
to the repo. Deal with it, they're small.

`make release` handles tagging and pushing to GitHub.

`make install` will perform the standard `go install` but will *also* install
the bash completions we make available for `dmk`. Note that this will use
`sudo` and it currently the only way to get the bash completions.

Before submiting a pull-request or merging into a mainline branch, you should
be sure that `make lint` passes with no errors. Originally we used `gometalinter`
with some customizations. Now we use `golangci-lint` with default settings. Use
`make lint-install` to install any tools necessary for linting.

Currently we do testing and linting in our CI pipeline via GitHub Actions.
