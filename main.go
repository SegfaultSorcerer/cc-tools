package main

import "cctools/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}