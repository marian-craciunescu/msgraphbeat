package main

import (
	"os"

	"github.com/marian-craciunescu/msgraphbeat/cmd"

	_ "github.com/marian-craciunescu/msgraphbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
