package main

import (
	"log"
)

// pcheck logs a detailed error and then panics with the same msg
func pcheck(err error) {
	if err != nil {
		log.Panicf("Fatal Error: %v\n", err)
	}
}
