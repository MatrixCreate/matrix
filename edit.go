package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func edit(cCtx *cli.Context) {
	var shallowMode = cCtx.Bool("shallow")
	var valetMode = cCtx.Bool("valet")

	projectName = cCtx.Args().First()

	if projectName == "" {
		color.Red("× Error: Missing project name")
		os.Exit(1)
	}

	if fileExists(projectName) {
		color.Red("× Error: Project directory already exists")
		os.Exit(2)
	}

	color.Magenta("Setting up existing project to edit: " + projectName)

	setupProject(false, shallowMode, valetMode)

	color.Magenta("Project Ready! cd " + projectName)
}
