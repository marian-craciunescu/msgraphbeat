package cmd

import (
	"github.com/marian-craciunescu/msgraphbeat/beater"

	cmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
)

// Name of this beat
var Name = "msgraphbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: Name})
