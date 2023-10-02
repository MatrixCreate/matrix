package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func create(cCtx *cli.Context) {
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

	color.Magenta("Creating new Craft CMS project: " + projectName)

	setupProject(true, false, valetMode)

	color.Magenta("Project Ready! cd " + projectName)
}
