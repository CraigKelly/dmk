// +build !test

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var buildDate string // Set by our build script

/////////////////////////////////////////////////////////////////////////////
// Entry point

var verb *log.Logger

func main() {
	log.SetFlags(0)
	log.Printf("dmk - built %s\n", buildDate)

	flags := flag.NewFlagSet("twivility", flag.ExitOnError)
	pipelineFile := flags.String("f", "Pipeline", "Pipeline file name")
	clean := flags.Bool("c", false, "Clean instead of build")
	verbose := flags.Bool("v", false, "verbose output")

	pcheck(flags.Parse(os.Args[1:]))

	// If it should always be printed, we use log. If it should only be printed
	// verbose=true, then we use verb
	if *verbose {
		verb = log.New(os.Stdout, "", 0)
	} else {
		verb = log.New(ioutil.Discard, "", 0)
	}

	verb.Printf("Verbose mode: ON\n")
	verb.Printf("Clean: %v\n", *clean)
	verb.Printf("Pipeline File: %s\n", *pipelineFile)

	// TODO: read config file
	cfgText, err := ioutil.ReadFile(*pipelineFile)
	pcheck(err)
	verb.Printf("Read %d bytes from %s\n", len(cfgText), *pipelineFile)

	cfg, err := ReadConfig(cfgText)
	pcheck(err)
	verb.Printf("Found %d build steps", len(cfg))

	if *clean {
		doClean(cfg)
	} else {
		doBuild(cfg)
	}
}

func doClean(cfg ConfigFile) {
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

	for _, file := range targetFiles {
		log.Printf("CLEAN: %s\n", file)
		// TODO: clean file
	}
}

func doBuild(cfg ConfigFile) {
	log.Printf("TODO: actually perform build\n")
}
