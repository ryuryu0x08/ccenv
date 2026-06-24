package main

import (
	"fmt"
	"os"

	"ccenv/cmd"
)

func main() {
	args := os.Args[1:]
	// If the first arg is a non-reserved, non-flag word, treat it as a profile
	// name and launch, passing through the rest. Everything else (reserved
	// subcommands, bare invocation, --help) goes through cobra.
	if len(args) > 0 && !cmd.IsReserved(args[0]) && args[0] != "" && args[0][0] != '-' {
		code, err := cmd.RunLaunch(args[0], args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "ccenv:", err)
		}
		os.Exit(code)
	}
	if err := cmd.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ccenv:", err)
		os.Exit(1)
	}
}
