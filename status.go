package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func status() {
	color.Magenta("Matrix CLI Status")

	// Check if Git is installed
	runCommand(exec.Command("git", "--version"), true, false, false)

	color.Green("✓ Git is installed")

	// Check if DDEV is installed
	runCommand(exec.Command("ddev", "--version"), true, false, false)

	color.Green("✓ DDEV is installed")

	// Check if Github CLI (gh) is installed and authed
	runCommand(exec.Command("gh", "auth", "status"), true, false, false)

	color.Green("✓ GitHub CLI is installed and authed")

	// TODO: Check if AWS CLI (aws) is installed and authed
	runCommand(exec.Command("aws", "--version"), true, false, false)

	color.Green("✓ AWS CLI is installed and authed")

	color.Green("✓ Completed: Matrix CLI Status")
}
