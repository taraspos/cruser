package main

import (
	"flag"

	"github.com/trane9991/cruser/core"
)

func main() {
	fileWithKeys := flag.String("file", "users", "File with the list of SSH-keys and user emails in format of '~/.ssh/authorized_keys' file")
	dryRun := flag.Bool("dry-run", false, "Do not execute commands, just print them.")
	flag.Parse()
	core.Run(fileWithKeys, dryRun)
}
