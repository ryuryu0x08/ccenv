package main

import (
	"fmt"
	"os"

	"ccenv/cmd"
)

func main() {
	args := os.Args[1:]
	// Strip ccenv's own leading flags (e.g. --verbose) before the profile name
	// so they are never forwarded to claude. Everything after the profile name
	// is passed through untouched.
	verbose, args := cmd.StripGlobalFlags(args)

	root := cmd.NewRoot()
	// First arg that is not a registered subcommand and not a flag is treated
	// as a profile name to launch, passing the rest through to claude.
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' && !cmd.IsReservedIn(root, args[0]) {
		code, err := cmd.RunLaunch(args[0], args[1:], verbose)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ccenv:", err)
		}
		os.Exit(code)
	}
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ccenv:", err)
		os.Exit(1)
	}
}
