package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	buildUnstarted = iota
	buildStarted   = iota
	buildExecuting = iota
	buildCompleted = iota
	buildFailed    = iota
)

// BuildStepInstance is a BuildStep executing
type BuildStepInstance struct {
	Step    *BuildStep
	Deps    []string
	State   int
	verb    *log.Logger
	decider Decider
	broad   *Broadcaster
}

// Can be our stand in below OR bytes.buffer
type stepOutput interface {
	io.Writer
	String() string
}

// Used in place of bytes.Buffer for direct output
type directOutput struct {
	dest io.Writer
}

func newDirectOutput(w io.Writer) *directOutput {
	return &directOutput{
		dest: w,
	}
}

func (do *directOutput) Write(p []byte) (int, error) {
	return do.dest.Write(p)
}

func (do *directOutput) String() string {
	return ""
}

// NewBuildStepInst creates an unstarted instance from the BuildStep
func NewBuildStepInst(step *BuildStep, allOutputs map[string]bool, verb *log.Logger, broad *Broadcaster) *BuildStepInstance {
	deps := make([]string, 0, len(step.Inputs))
	for _, file := range step.Inputs {
		if _, inMap := allOutputs[file]; inMap {
			deps = append(deps, file)
		}
	}

	verb.Printf("%s: Found %d deps\n", step.Name, len(deps))

	// If verbose, output vars in sorted order
	varKeys := []string{}
	for k := range step.Vars {
		varKeys = append(varKeys, k)
	}
	sort.Strings(varKeys)
	for _, k := range varKeys {
		verb.Printf("%s: var[%s]=='%s'\n", step.Name, k, step.Vars[k])
	}

	return &BuildStepInstance{
		Step:    step,
		Deps:    deps,
		State:   buildUnstarted,
		verb:    verb,
		decider: TimeDecider{},
		broad:   broad,
	}
}

// Tell everyone that our outputs are done (even if we failed)
func (i *BuildStepInstance) notify() {
	for _, file := range i.Step.Outputs {
		i.verb.Printf("%s: notifying for %s\n", i.Step.Name, file)
		err := i.broad.Send(file)
		if err != nil {
			i.verb.Printf("%s: ERROR on broadcast send for %s - %v\n", i.Step.Name, file, err)
		}
	}
}

func (i *BuildStepInstance) fail(err error) error {
	i.notify()
	i.State = buildFailed
	log.Printf("%s: FAIL - %s\n", i.Step.Name, err.Error())
	return err
}

func (i *BuildStepInstance) succeed() error {
	i.notify()
	i.State = buildCompleted
	log.Printf("%s: Complete\n", i.Step.Name)
	return nil
}

// Run actually executes the build command properly
// Note that this function should call .succeed or .fail before exiting
func (i *BuildStepInstance) Run() error {
	// In case we somehow don't correctly leave
	defer func() {
		if i.State == buildFailed || i.State == buildCompleted {
			return
		}
		// We failed to complete a step
		log.Printf("%s: FAILURE TO CLEANLY FINISH STATE - this is a bug\n", i.Step.Name)
		err := i.fail(errors.New("Build step instance state indeterminate"))
		if err != nil {
			i.verb.Printf("And could not fail the step! %v\n", err)
		}
	}()

	// The step is "Started"
	i.State = buildStarted

	// If any of the required inputs are another step's outputs, then wait for
	// a built message for all our deps
	if len(i.Deps) > 0 {
		waitingDeps := make(map[string]bool)
		for _, d := range i.Deps {
			waitingDeps[d] = true
		}
		i.verb.Printf("%s: waiting for %d deps\n", i.Step.Name, len(waitingDeps))

		list := i.broad.GetListener()

		for msg := range list.Delivery {
			file := msg.Msg
			delete(waitingDeps, file)
			if len(waitingDeps) > 0 {
				list.Respond(true) // Keep working
			} else {
				list.Respond(false) // Have everything we need!
				i.verb.Printf("%s: all deps are done - proceeding\n", i.Step.Name)
				break
			}
		}

		// IF we still have deps, then something has gone really wrong
		if len(waitingDeps) > 0 {
			i.verb.Printf("%s: %d deps will never finish, exiting\n", i.Step.Name, len(waitingDeps))
			return i.fail(fmt.Errorf("Broadcaster done, still have %d deps", len(waitingDeps)))
		}
	}

	// If we have inputs, check to see if we need to build
	needBuild, err := i.decider.NeedBuild(i.Step.Inputs, i.Step.Outputs)
	if err != nil {
		i.verb.Printf("%s: failing on build decision\n", i.Step.Name)
		return i.fail(err)
	}
	if !needBuild {
		i.verb.Printf("%s: Nothing to do\n", i.Step.Name)
		return i.succeed()
	}

	// Time to execute!
	i.State = buildExecuting
	log.Printf("%s: %s\n", i.Step.Name, i.Step.Command)

	cmd := exec.Command("/bin/bash", "-c", i.Step.Command)

	env := os.Environ()
	// Some variables are already set in our environment
	// DMK_VERSION is set on startup, DMK_PIPELINE is set after reading the file
	env = append(env, fmt.Sprintf("DMK_STEPNAME=%s", i.Step.Name))
	env = append(env, fmt.Sprintf("DMK_INPUTS=%v", strings.Join(i.Step.Inputs, ":")))
	env = append(env, fmt.Sprintf("DMK_OUTPUTS=%v", strings.Join(i.Step.Outputs, ":")))
	env = append(env, fmt.Sprintf("DMK_CLEAN=%v", strings.Join(i.Step.Clean, ":")))
	// Now add the step variables to the environment
	for k, v := range i.Step.Vars {
		env = append(env, fmt.Sprintf("%v=%v", k, v))
	}
	cmd.Env = env

	var stdOut stepOutput
	var stdErr stepOutput

	if i.Step.Direct {
		stdOut = newDirectOutput(os.Stdout)
		stdErr = newDirectOutput(os.Stderr)
	} else {
		stdOut = &bytes.Buffer{}
		stdErr = &bytes.Buffer{}
	}

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	cmdErr := cmd.Run()

	stdoutText := strings.TrimSpace(stdOut.String())
	stderrText := strings.TrimSpace(stdErr.String())
	if len(stdoutText) > 0 {
		i.verb.Printf("%s stdout begin---\n%s\n---stdout end for %s\n",
			i.Step.Name,
			stdOut.String(),
			i.Step.Name)
	}
	if len(stderrText) > 0 {
		log.Printf("%s stderr begin---\n%s\n---stderr end for %s\n",
			i.Step.Name,
			stdErr.String(),
			i.Step.Name)
	}

	if cmdErr != nil {
		return i.fail(cmdErr)
	}

	// If we still need a build, then we failed
	stillNeedBuild, err := i.decider.NeedBuild(i.Step.Inputs, i.Step.Outputs)
	if err != nil {
		return i.fail(fmt.Errorf("Build decider check failed AFTER build: %s", err.Error()))
	}
	if stillNeedBuild {
		return i.fail(errors.New("Build still required after command finished"))
	}

	// if any outputs missing return failed
	for _, file := range i.Step.Outputs {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return i.fail(err)
		}
	}

	return i.succeed()
}
