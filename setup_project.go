package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func setupProject(freshMode bool, shallowMode bool, valetMode bool) {
	if freshMode {
		// git clone --depth=1 {craftStarterRepo} {projectName}
		runCommand(exec.Command("git", "clone", "--depth=1", craftStarterRepo, projectName), false, false, true)

		// ddev config --project-name={projectName}
		runCommand(exec.Command("ddev", "config", "--project-name="+projectName), false, true, false)
	} else {
		if shallowMode {
			// git clone --depth=1 --no-single-branch -b develop git@bitbucket.org:matrixcreate/{projectName}.git {projectName}
			runCommand(exec.Command("git", "clone", "--depth=1", "--no-single-branch", "-b", "develop", "git@bitbucket.org:matrixcreate/"+projectName+".git", projectName), false, false, true)
		} else {
			// git clone -b develop git@bitbucket.org:matrixcreate/{projectName}.git {projectName}
			runCommand(exec.Command("git", "clone", "-b", "develop", "git@bitbucket.org:matrixcreate/"+projectName+".git", projectName), false, false, true)
		}
	}

	if valetMode {
		// valet link
		runCommand(exec.Command("valet", "link"), false, true, true)

		// composer install
		if fileExists(projectName + "/composer.lock") {
			runCommand(exec.Command("composer", "install"), false, true, false)
		} else {
			color.Yellow("- No composer.lock file found. Skipping composer install")
		}

		// npm install
		if fileExists(projectName + "/package-lock.json") {
			runCommand(exec.Command("npm", "install"), false, true, false)
		} else {
			color.Yellow("- No package-lock.json file found. Skipping npm install")
		}

		if fileExists(projectName + "/craft") {
			// php craft setup/app-id --interactive=0
			runCommand(exec.Command("php", "craft", "setup/app-id", "--interactive=0"), false, true, false)

			// php craft setup/security-key
			runCommand(exec.Command("php", "craft", "setup/security-key"), false, true, false)

			// php craft setup/db --interactive=0 --driver=mysql --database=db --password=db --user=db --server=ddev-{projectName}-db --port=3306
			runCommand(exec.Command("php", "craft", "setup/db", "--interactive=0", "--driver=mysql", "--database=db", "--password=db", "--user=db", "--server=ddev-"+projectName+"-db", "--port=3306"), false, true, false)
		}

		// ddev import-db --src=_db/db.zip
		// TODO: Add Valet version for DB settings using DBngin

		if freshMode {
			// rm -rf ./{projectName}/.git
			runCommand(exec.Command("rm", "-rf", "./"+projectName+"/.git"), false, true, false)

			// git init
			runCommand(exec.Command("git", "init"), false, true, false)
		}
	} else {
		// ddev start
		runCommand(exec.Command("ddev", "start"), false, true, true)

		// ddev composer install
		if fileExists(projectName + "/composer.lock") {
			runCommand(exec.Command("ddev", "composer", "install"), false, true, false)
		} else {
			color.Yellow("- No composer.lock file found. Skipping composer install")
		}

		// ddev npm install
		if fileExists(projectName + "/package-lock.json") {
			runCommand(exec.Command("ddev", "npm", "install"), false, true, false)
		} else {
			color.Yellow("- No package-lock.json file found. Skipping npm install")
		}

		if fileExists(projectName + "/craft") {
			// ddev craft setup/app-id --interactive=0
			runCommand(exec.Command("ddev", "craft", "setup/app-id", "--interactive=0"), false, true, false)

			// ddev craft setup/security-key
			runCommand(exec.Command("ddev", "craft", "setup/security-key"), false, true, false)

			// ddev craft setup/db --interactive=0 --driver=mysql --database=db --password=db --user=db --server=ddev-{projectName}-db --port=3306
			runCommand(exec.Command("ddev", "craft", "setup/db", "--interactive=0", "--driver=mysql", "--database=db", "--password=db", "--user=db", "--server=ddev-"+projectName+"-db", "--port=3306"), false, true, false)
		}

		// ddev import-db --src=_db/db.zip
		if fileExists(projectName + "/_db/db.zip") {
			runCommand(exec.Command("ddev", "import-db", "--src=_db/db.zip"), false, true, false)
		} else {
			color.Yellow("- No _db/db.zip file found. Skipping ddev import-db")
		}

		if freshMode {
			// rm -rf ./{projectName}/.git
			runCommand(exec.Command("rm", "-rf", "./"+projectName+"/.git"), false, true, false)

			// git init
			runCommand(exec.Command("git", "init"), false, true, false)
		}

		// ddev describe
		runCommand(exec.Command("ddev", "describe"), true, true, false)
	}
}
