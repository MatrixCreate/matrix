package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func create(cCtx *cli.Context) {
	ProjectName = cCtx.Args().First()

	if ProjectName == "" {
		color.Red("× Error: Missing project name")
		os.Exit(1)
	}

	if fileExists(ProjectName) {
		color.Red("× Error: Project directory already exists")
		os.Exit(2)
	}

	color.Magenta("Creating new Craft CMS project: " + ProjectName)

	setupProject(true, false)

	color.Magenta("Project Ready! cd " + ProjectName)
}
