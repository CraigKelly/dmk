package main

import (
	"errors"
	"log"
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
}

// NewBuildStepInst creates an unstarted instance from the BuildStep
func NewBuildStepInst(step *BuildStep, allOutputs map[string]bool, verb *log.Logger) *BuildStepInstance {
	deps := make([]string, 0, len(step.Inputs))
	for _, file := range step.Inputs {
		if _, inMap := allOutputs[file]; inMap {
			deps = append(deps, file)
		}
	}

	verb.Printf("%s: Found %d deps\n", step.Name, len(deps))

	return &BuildStepInstance{
		Step:    step,
		Deps:    deps,
		State:   buildUnstarted,
		verb:    verb,
		decider: TimeDecider{},
	}
}

// Tell everyone that our outputs are done (even if we failed)
func (i *BuildStepInstance) notify() {
	for _, file := range i.Step.Outputs {
		// TODO: actual broadcast
		i.verb.Printf("%s: notifying for %s\n", i.Step.Name, file)
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
		i.fail(errors.New("Build step instance state indeterminate"))
	}()

	// The step is "Started"
	i.State = buildStarted

	// If any of the required inputs are another step's outputs, then wait for
	// a built message.
	if len(i.Deps) > 0 {
		// TODO: wait for built message(s)
	}

	// If we have inputs, check to see if we need to build
	needBuild, err := i.decider.NeedBuild(i.Step.Inputs, i.Step.Outputs)
	if err != nil {
		return i.fail(err)
	}
	if !needBuild {
		i.verb.Printf("%s: Nothing to do\n", i.Step.Name)
		return i.succeed()
	}

	// Time to execute!
	i.State = buildExecuting

	// TODO: run the command

	// TODO: if command retcode was nonzero return failed
	// TODO: if any outputs missing return failed
	// TODO: if any outputs older than inputs, then return failed

	return i.succeed()
}
