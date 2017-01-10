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

	verb.Printf("Verbose mode ON\n")
	verb.Printf("Clean == %v\n", *clean)
	verb.Printf("Pipeline File == %s\n", *pipelineFile)

	// TODO: read config file

	if *clean {
		doClean()
	} else {
		doBuild()
	}
}

func doClean() {
	log.Printf("TODO: actually perform clean\n")
}

func doBuild() {
	log.Printf("TODO: actually perform build\n")
}
