package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func runCommand(cmd *exec.Cmd, showOutput bool, inProject bool, exitOnError bool) {
	s.Start()

	if inProject {
		cmd.Dir = "./" + projectName
	}

	color.White("Running: " + cmd.String())

	if showOutput {
		out, err := cmd.Output()
		if err != nil {
			s.Stop()

			if exitOnError {
				color.Red("× Error Running: " + cmd.String())
				color.Red("× " + err.Error())
				os.Exit(commandCount)
			} else {
				color.Yellow("× Error Running: " + cmd.String())
				color.Yellow("× " + err.Error())
			}
		}
		fmt.Println(string(out))
	} else {
		err := cmd.Run()
		if err != nil {
			s.Stop()

			if exitOnError {
				color.Red("× Error Running: " + cmd.String())
				color.Red("× " + err.Error())
				color.White("Tip: Run the above command separately for more info to find out what went wrong")
				os.Exit(commandCount)
			} else {
				color.Yellow("× Error Running: " + cmd.String())
				color.Yellow(err.Error())
			}
		} else {
			s.Stop()
			color.Green("✓ Completed: " + cmd.String())
		}
	}

	commandCount++
}

func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	} else {
		return false
	}
}
