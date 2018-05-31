package main

import (
	cmd "github.com/go-fusion/cmd/fusiond/commands"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
