package main

import (
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func delete(cCtx *cli.Context) {
	color.White(ProjectName)

	ProjectName = cCtx.Args().First()

	if ProjectName == "" {
		color.Red("× Error: Missing project name")
		os.Exit(1)
	}

	if !fileExists(ProjectName) {
		color.Red("× Error: Project directory not found")
		os.Exit(2)
	}

	color.Magenta("Deleting project: " + ProjectName)

	// ddev stop --remove-data --omit-snapshot
	runCommand(exec.Command("ddev", "stop", "--remove-data", "--omit-snapshot"), false, true, false)

	// rm -rf {ProjectName}
	runCommand(exec.Command("rm", "-rf", ProjectName), false, false, true)

	color.Magenta("Project Deleted!")
}
