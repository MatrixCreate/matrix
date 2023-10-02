package main

import (
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func delete(cCtx *cli.Context) {
	color.White(projectName)

	projectName = cCtx.Args().First()

	if projectName == "" {
		color.Red("× Error: Missing project name")
		os.Exit(1)
	}

	if !fileExists(projectName) {
		color.Red("× Error: Project directory not found")
		os.Exit(2)
	}

	color.Magenta("Deleting project: " + projectName)

	// ddev stop --remove-data --omit-snapshot
	runCommand(exec.Command("ddev", "stop", "--remove-data", "--omit-snapshot"), false, true, false)

	// rm -rf {projectName}
	runCommand(exec.Command("rm", "-rf", projectName), false, false, true)

	color.Magenta("Project Deleted!")
}
