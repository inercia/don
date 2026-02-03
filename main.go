// Package main is the entry point for the don CLI application.
//
// don is an AI agent that connects Large Language Models to command-line tools.
// It provides direct LLM connectivity without requiring a separate MCP client.
package main

import (
	root "github.com/inercia/don/cmd"
)

// Version information set by build flags
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	root.SetVersion(version, commit, buildDate)
	root.Execute()
}
