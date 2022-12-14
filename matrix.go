package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var craftStarterRepo string = "git@github.com:MatrixCreate/craft-starter.git"
var projectName string = ""
var commandCount int = 0

func main() {
	setupEnv()

	app := &cli.App{
		Name:      "Matrix CLI",
		Version:   "v1.2.0",
		Copyright: "(c) 2023 Matrix Create",
		Usage:     "Project Management CLI Tool",
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new Craft CMS project",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "valet",
						Aliases: []string{"v"},
						Usage:   "Edit using Valet instead of DDEV",
					},
				},
				Action: func(cCtx *cli.Context) error {
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

					return nil
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Clone and setup an existing project to edit",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "shallow",
						Aliases: []string{"s"},
						Usage:   "Edit in shallow mode which provides a low depth git clone with all branches",
					},
					&cli.BoolFlag{
						Name:    "valet",
						Aliases: []string{"v"},
						Usage:   "Edit using Valet instead of DDEV",
					},
				},
				Action: func(cCtx *cli.Context) error {
					var shallowMode = cCtx.Bool("shallow")
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

					color.Magenta("Setting up existing project to edit: " + projectName)

					setupProject(false, shallowMode, valetMode)

					color.Magenta("Project Ready! cd " + projectName)

					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "Stop and delete project",
				Action: func(cCtx *cli.Context) error {
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

					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"self-update"},
				Usage:   "Self Update Matrix CLI",
				Action: func(cCtx *cli.Context) error {
					color.Magenta("Self Updating Matrix CLI")

					runCommand(exec.Command("go", "install", "github.com/MatrixCreate/matrix@latest"), false, false, true)

					color.Magenta("Updated!")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupEnv() {
	if fileExists(".env") {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

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

func runCommand(cmd *exec.Cmd, showOutput bool, inProject bool, exitOnError bool) {
	if inProject {
		cmd.Dir = "./" + projectName
	}

	fmt.Println("Running: " + cmd.String())

	if showOutput {
		out, err := cmd.Output()
		if err != nil {
			if exitOnError {
				color.Red("× Error Running: " + cmd.String())
				color.Red(err.Error())
				os.Exit(commandCount)
			} else {
				color.Yellow("× Error Running: " + cmd.String())
				color.Yellow(err.Error())
			}
		}
		fmt.Println(string(out))
	} else {
		err := cmd.Run()
		if err != nil {
			if exitOnError {
				color.Red("× Error Running: " + cmd.String())
				color.Red(err.Error())
				os.Exit(commandCount)
			} else {
				color.Yellow("× Error Running: " + cmd.String())
				color.Yellow(err.Error())
			}
		} else {
			color.Green("✓ Completed: " + cmd.String())
		}
	}

	commandCount++
}

func runRemoteCommand(command string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		color.Red("× Error: Unable to find home dir: %v", err)
		os.Exit(1)
	}
	key, err := os.ReadFile(homeDir + "/.ssh/id_rsa")
	if err != nil {
		color.Red("× Error: Unable to read private key: %v", err)
		os.Exit(2)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		color.Red("× Error: Unable to parse private key: %v", err)
		os.Exit(3)
	}
	hostKeyCallback, err := knownhosts.New(homeDir + "/.ssh/known_hosts")
	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User: os.Getenv("REMOTE_SERVER_USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}
	client, err := ssh.Dial("tcp", os.Getenv("REMOTE_SERVER_IP")+":22", config)
	if err != nil {
		log.Fatal("× Failed to dial: ", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("× Failed to create session: ", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("× Failed to run: " + err.Error())
	}

	color.White(b.String())
}

func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	} else {
		return false
	}
}
