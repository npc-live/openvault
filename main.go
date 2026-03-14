package main

import "github.com/npc-live/openvault/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
