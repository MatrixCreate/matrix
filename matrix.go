package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var craftStarterRepo = "git@github.com:MatrixCreate/craft-starter.git"
var projectName = ""

func main() {
	app := &cli.App{
		Name:      "Matrix CLI",
		Version:   "v1.0",
		Copyright: "(c) 2022 Matrix Create",
		Usage:     "Project Management CLI Tool",
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new Craft CMS project",
				Action: func(cCtx *cli.Context) error {
					projectName = cCtx.Args().First()

					if projectName == "" {
						color.Red("Missing project name")
						return nil
					}

					color.Magenta("Creating new Craft CMS project: " + projectName)

					// git clone --depth=1 {craftStarterRepo} {projectName}
					cmd := exec.Command("git", "clone", "--depth=1", craftStarterRepo, projectName)
					cmdErr := cmd.Run()
					if cmdErr != nil {
						if cmdErr.Error() == "exit status 128" {
							color.Red("Project already exists")
							return nil
						}

						color.Red("Error (git clone): " + cmdErr.Error())
						return nil
					}

					color.Green("✓ git clone craft-starter")

					setupCraftCMS(true)

					// ddev describe
					runCommand(exec.Command("ddev", "describe"), true)

					color.Magenta("Project Ready! /" + projectName)

					return nil
				},
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy project from current directory",
				Action: func(cCtx *cli.Context) error {
					// TODO:
					color.Red("TODO: Command not ready yet")

					return nil
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Clone and setup an existing project to edit",
				Action: func(cCtx *cli.Context) error {
					projectName = cCtx.Args().First()

					if projectName == "" {
						color.Red("Missing project name")
						return nil
					}

					color.Magenta("Editing existing Craft CMS project: " + projectName)

					// git clone --depth=1 {craftStarterRepo} {projectName}
					cmd := exec.Command("git", "clone", "--depth=1", "git@bitbucket.org:matrixcreate/"+projectName+".git", projectName)
					cmdErr := cmd.Run()
					if cmdErr != nil {
						if cmdErr.Error() == "exit status 128" {
							color.Red("Project already exists")
							return nil
						}

						color.Red("Error (git clone): " + cmdErr.Error())
						return nil
					}

					setupCraftCMS(false)

					// ddev describe
					runCommand(exec.Command("ddev", "describe"), true)

					color.Magenta("Project Ready! /" + projectName)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupCraftCMS(fresh bool) {
	if fresh {
		// ddev config --project-name={projectName}
		runCommand(exec.Command("ddev", "config", "--project-name="+projectName), false)
		color.Green("✓ ddev config --project-name=" + projectName)
	}

	// composer install
	runCommand(exec.Command("composer", "install"), false)
	color.Green("✓ composer install")

	// npm install
	runCommand(exec.Command("npm", "install"), false)
	color.Green("✓ npm install")

	// ddev start
	runCommand(exec.Command("ddev", "start"), false)
	color.Green("✓ ddev start")

	// ddev craft setup/app-id
	runCommand(exec.Command("ddev", "craft", "setup/app-id"), false)
	color.Green("✓ ddev craft setup/app-id")

	// ddev craft setup/security-key
	runCommand(exec.Command("ddev", "craft", "setup/security-key"), false)
	color.Green("✓ ddev craft setup/security-key")

	// ddev craft setup/db --interactive=0 --driver=mysql --database=db --password=db --user=db --server=ddev-{projectName}-db --port=3306
	runCommand(exec.Command("ddev", "craft", "setup/db", "--interactive=0", "--driver=mysql", "--database=db", "--password=db", "--user=db", "--server=ddev-"+projectName+"-db", "--port=3306"), false)
	color.Green("✓ ddev craft setup/db")

	// ddev import-db --src=_db/db.zip
	runCommand(exec.Command("ddev", "import-db", "--src=_db/db.zip"), false)
	color.Green("✓ ddev import-db")

	if fresh {
		// rm -rf ./{projectName}/.git
		runCommand(exec.Command("rm", "-rf", "./"+projectName+"/.git"), false)
		color.Green("✓ rm -rf ./" + projectName + "/.git")

		// git init
		runCommand(exec.Command("git", "init"), false)
		color.Green("✓ git init")
	}
}

func runCommand(cmd *exec.Cmd, showOutput bool) {
	cmd.Dir = "./" + projectName

	if showOutput {
		out, err := cmd.Output()
		if err != nil {
			color.Red("Error: " + err.Error())
		}
		fmt.Println(string(out))
	} else {
		err := cmd.Run()
		if err != nil {
			color.Red("Error: " + err.Error())
		}
	}
}
