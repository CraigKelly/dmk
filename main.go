package main

// TODO: -listSteps should NOT include abstract steps but it SHOULD include explicit steps

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

/////////////////////////////////////////////////////////////////////////////
// Entry point

func main() {
	log.SetFlags(0)
	pcheck(os.Setenv("DMK_VERSION", Version()))

	flags := flag.NewFlagSet("dmk", flag.ExitOnError)
	pipelineFileSpec := flags.String("f", "", "Pipeline file name")
	cleanSpec := flags.Bool("c", false, "Clean instead of build")
	verboseSpec := flags.Bool("v", false, "verbose output")
	listStepsSpec := flags.Bool("listSteps", false, "list all steps and exit. No other actions will be taken")

	pcheck(flags.Parse(os.Args[1:]))

	clean := *cleanSpec
	verbose := *verboseSpec
	args := flags.Args()
	listSteps := *listStepsSpec

	if !listSteps {
		log.Printf("dmk %s\n", Version())
	}

	// If they didn't select a pipeline file, we try to find a default
	var pipelineFile string
	if pipelineFileSpec == nil || *pipelineFileSpec == "" {
		pipelineFile = FirstFileFound(
			"Pipeline", "Pipeline.yaml",
			"pipeline", "pipeline.yaml",
			".Pipeline", ".Pipeline.yaml",
			".pipeline", ".pipeline.yaml",
		)
		if pipelineFile == "" {
			pipelineFile = "Pipeline.yaml" // choose what we'll show
		}
	} else {
		pipelineFile = *pipelineFileSpec
	}

	// If it should always be printed, we use log. If it should only be printed
	// verbose=true, then we use verb
	var verb *log.Logger
	if verbose {
		verb = log.New(os.Stdout, "", 0)
	} else {
		verb = log.New(ioutil.Discard, "", 0)
	}

	verb.Printf("Verbose mode: ON\n")
	verb.Printf("Clean: %v\n", clean)
	verb.Printf("Pipeline File: %s\n", pipelineFile)
	verb.Printf("List Steps: %v\n", listSteps)

	// read the config file
	cfgText, err := ioutil.ReadFile(pipelineFile)
	if os.IsNotExist(err) {
		if !listSteps {
			log.Printf("%s does not exist - exiting\n", pipelineFile)
		}
		return
	}
	pcheck(err)
	verb.Printf("Read %d bytes from %s\n", len(cfgText), pipelineFile)

	// Before we change directory, go ahead save the absolute path of the
	// pipeline file in the environment
	absPipelineFile, err := filepath.Abs(pipelineFile)
	pcheck(err)
	pcheck(os.Setenv("DMK_PIPELINE", absPipelineFile))

	// Change to the pipeline file's directory: note that this must happen
	// before we parse the config file for globbing to work
	pipelineDir := filepath.Dir(pipelineFile)
	if pipelineDir != "." {
		verb.Printf("Changing current directory to: %s\n", pipelineDir)
	}
	pcheck(os.Chdir(pipelineDir))

	// Parse the config file
	cfg, err := ReadConfig(cfgText)
	pcheck(err)
	verb.Printf("Found %d build steps", len(cfg))

	// Figure out the steps that need to run
	var newCfg ConfigFile
	if len(args) > 0 {
		verb.Printf("Steps specified on command line: trimming for %v\n", args)
		newCfg, err = TrimSteps(cfg, args)
		pcheck(err)
		cfg = newCfg
		verb.Printf("%d build steps remaining", len(cfg))
	} else {
		verb.Printf("No steps specified: removing steps where explicit=true\n")
		newCfg, err = NoExplicit(cfg)
		pcheck(err)
		cfg = newCfg
		verb.Printf("%d build steps remaining", len(cfg))
	}

	// Do what we're supposed to do
	var exitCode int
	if listSteps {
		exitCode = DoListSteps(cfg, verb)
	} else if clean {
		exitCode = DoClean(cfg, verb)
	} else {
		exitCode = DoBuild(cfg, verb)
	}

	os.Exit(exitCode)
}

// DoListSteps just outputs all step names
func DoListSteps(cfg ConfigFile, verb *log.Logger) int {
	// We must write to stdout, so we always create our own logger
	stepLog := log.New(os.Stdout, "", 0)
	for _, step := range cfg {
		stepLog.Printf("%s\n", step.Name)
	}
	return 0
}

// DoClean cleans all files specified by the config file
func DoClean(cfg ConfigFile, verb *log.Logger) int {
	targets := NewUniqueStrings()

	for _, step := range cfg {
		for _, file := range step.Outputs {
			targets.Add(file)
		}
		for _, file := range step.Clean {
			targets.Add(file)
		}
	}

	targetFiles := targets.Strings()
	verb.Printf("Cleaning %d files\n", len(targetFiles))

	failCount := 0
	for _, file := range targetFiles {
		log.Printf("CLEAN: %s\n", file)
		err := os.RemoveAll(file)
		if err != nil && !os.IsNotExist(err) {
			failCount++
			log.Printf("  failed to clean: %s\n", err.Error())
		}
	}

	return failCount
}

// DoBuild um, does the build
func DoBuild(cfg ConfigFile, verb *log.Logger) int {
	// Get all targets (outputs)
	targets := NewUniqueStrings()
	for _, step := range cfg {
		for _, file := range step.Outputs {
			targets.Add(file)
		}
	}
	verb.Printf("BUILD: total possible outputs = %d\n", len(targets.Seen))

	// We need a broadcaster for dependency notifications
	broad := NewBroadcaster()
	pcheck(broad.Start())

	// Start all steps running
	running := make([]*BuildStepInstance, 0, len(cfg))
	wg := sync.WaitGroup{}

	for _, step := range cfg {
		verb.Printf("Starting step %s\n", step.Name)

		one := NewBuildStepInst(step, targets.Seen, verb, broad)
		running = append(running, one)

		wg.Add(1)
		go func(inst *BuildStepInstance) {
			defer wg.Done()
			err := inst.Run()
			if err != nil {
				verb.Printf("%s: %s\n", inst.Step.Name, err.Error())
			}
		}(one)
	}

	// Wait for them to complete
	wg.Wait()
	broad.Kill()

	// Determine and use exit code
	failCount := 0
	successCount := 0
	for _, step := range running {
		if step.State == buildCompleted {
			successCount++
		} else if step.State == buildFailed {
			failCount++
			DeleteFailed(step.Step) // Remove any outputs on fail
		}
	}
	if failCount+successCount < len(running) {
		log.Printf("Fatal error: at least one step has NOT completed\n")
		failCount = failCount + successCount + 1
	}
	return failCount
}

// DeleteFailed deletes the output for a failed step if necessary
func DeleteFailed(step *BuildStep) {
	if !step.DelOnFail {
		return
	}
	for _, f := range step.Outputs {
		err := os.RemoveAll(f)
		if err == nil || !os.IsNotExist(err) {
			log.Printf("%s: deleted %s\n", step.Name, f)
		} else if err != nil {
			log.Printf("%s: tried to delete %s but failed: %s\n", step.Name, f, err.Error())
		}
	}
}
