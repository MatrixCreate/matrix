package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func update() {
	color.Magenta("Self Updating Matrix CLI")

	runCommand(exec.Command("go", "install", "github.com/MatrixCreate/matrix@latest"), false, false, true)

	color.Magenta("--------------------------------------------------")
	color.Magenta("🎉            UPDATE COMPLETE                   🎉")
	color.Magenta("--------------------------------------------------")
}
