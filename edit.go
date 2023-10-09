package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func edit(cCtx *cli.Context) {
	var shallowMode = cCtx.Bool("shallow")
	var valetMode = cCtx.Bool("valet")

	ProjectName = cCtx.Args().First()

	if ProjectName == "" {
		color.Red("× Error: Missing project name")
		os.Exit(1)
	}

	if fileExists(ProjectName) {
		color.Red("× Error: Project directory already exists")
		os.Exit(2)
	}

	color.Magenta("Setting up existing project to edit: " + ProjectName)

	setupProject(false, shallowMode, valetMode)

	color.Magenta("Project Ready! cd " + ProjectName)
}
