package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func setupProject(freshMode bool, shallowMode bool) {
	if freshMode {
		// git clone --depth=1 {CraftStarterRepo} {ProjectName}
		runCommand(exec.Command("git", "clone", "--depth=1", CraftStarterRepo, ProjectName), false, false, true)

		// ddev config --project-name={ProjectName}
		runCommand(exec.Command("ddev", "config", "--project-name="+ProjectName), false, true, false)
	} else {
		if shallowMode {
			// git clone --depth=1 --no-single-branch -b develop git@github.com:{GithubRepoUser}/{ProjectName}.git {ProjectName}
			runCommand(exec.Command("git", "clone", "--depth=1", "--no-single-branch", "-b", "develop", "git@github.com:"+GithubRepoUser+"/"+ProjectName+".git", ProjectName), false, false, false)
		} else {
			// git clone -b develop git@git/usr/bin/git clone -b develop git@github.com:matrixcreate/beau-bronzage-laravel.git beau-bronzage-laravelhub.com:{GithubRepoUser}/{ProjectName}.git {ProjectName}
			runCommand(exec.Command("git", "clone", "-b", "develop", "git@github.com:"+GithubRepoUser+"/"+ProjectName+".git", ProjectName), false, false, false)
		}

		// Check if that worked
		if !fileExists(ProjectName) {
			color.Yellow("? Maybe develop branch dosn't exist, trying main branch...")

			// Try in main branch
			if shallowMode {
				// git clone --depth=1 --no-single-branch git@github:{GithubRepoUser}/{ProjectName} {ProjectName}
				runCommand(exec.Command("git", "clone", "--depth=1", "--no-single-branch", "git@github.com:"+GithubRepoUser+"/"+ProjectName+".git", ProjectName), false, false, true)
			} else {
				// git clone git@github:{GithubRepoUser}/{ProjectName} {ProjectName}
				runCommand(exec.Command("git", "clone", "git@github.com:"+GithubRepoUser+"/"+ProjectName+".git", ProjectName), false, false, true)
			}
		}
	}

	// ddev start
	if fileExists(ProjectName + "/.ddev") {
		runCommand(exec.Command("ddev", "start"), false, true, true)
	}

	// ddev composer install
	if fileExists(ProjectName + "/composer.lock") {
		runCommand(exec.Command("ddev", "composer", "install"), false, true, false)
	} else {
		color.Yellow("- No composer.lock file found. Skipping composer install")
	}

	// ddev npm install
	if fileExists(ProjectName + "/package-lock.json") {
		// For Laravel we run npm locally
		if fileExists(ProjectName+"/artisan") || !fileExists(ProjectName+"/.ddev") {
			runCommand(exec.Command("npm", "install"), false, true, false)
		} else {
			runCommand(exec.Command("ddev", "npm", "install"), false, true, false)
		}
	} else {
		color.Yellow("- No package-lock.json file found. Skipping npm install")
	}

	if fileExists(ProjectName + "/craft") {
		// ddev craft setup/app-id --interactive=0
		runCommand(exec.Command("ddev", "craft", "setup/app-id", "--interactive=0"), false, true, false)

		// ddev craft setup/security-key
		runCommand(exec.Command("ddev", "craft", "setup/security-key"), false, true, false)

		// ddev craft setup/db --interactive=0 --driver=mysql --database=db --password=db --user=db --server=ddev-{ProjectName}-db --port=3306
		runCommand(exec.Command("ddev", "craft", "setup/db", "--interactive=0", "--driver=mysql", "--database=db", "--password=db", "--user=db", "--server=ddev-"+ProjectName+"-db", "--port=3306"), false, true, false)
	}

	// ddev import-db --file=_db/db.zip
	if fileExists(ProjectName + "/_db/db.zip") {
		runCommand(exec.Command("ddev", "import-db", "--file=_db/db.zip"), false, true, false)
	} else {
		color.Yellow("- No _db/db.zip file found. Skipping ddev import-db")
	}

	// If Laravel then ddev artisan migrate --seed and key:generate
	if fileExists(ProjectName + "/artisan") {
		runCommand(exec.Command("ddev", "artisan", "migrate", "--seed"), false, true, false)
		runCommand(exec.Command("ddev", "artisan", "key:generate"), false, true, false)
		runCommand(exec.Command("npm", "run", "build"), false, true, false)
	} else {
		color.Yellow("- No artisan file found. Skipping ddev artisan migrate")
	}

	if freshMode {
		// rm -rf ./{ProjectName}/.git
		runCommand(exec.Command("rm", "-rf", "./"+ProjectName+"/.git"), false, true, false)

		// git init
		runCommand(exec.Command("git", "init"), false, true, false)
	}

	// ddev describe
	runCommand(exec.Command("ddev", "describe"), true, true, false)
}
