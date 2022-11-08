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

func main() {
	setupEnv()

	app := &cli.App{
		Name:      "Matrix CLI",
		Version:   "v1.0.6",
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

					if fileExists(projectName) {
						color.Red("Project directory already exists")
						os.Exit(0)
					}

					color.Magenta("Creating new Craft CMS project: " + projectName)

					setupCraftCMS(true)

					color.Magenta("Project Ready! /" + projectName)

					return nil
				},
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy project from current directory",
				Action: func(cCtx *cli.Context) error {
					color.Red("This command is not finished yet!")

					runRemoteCommand("ls -l")

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

					color.Magenta("Setting up existing Craft CMS project to edit: " + projectName)

					setupCraftCMS(false)

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

func setupEnv() {
	if fileExists(".env") {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func setupCraftCMS(fresh bool) {
	if fresh {
		// git clone --depth=1 {craftStarterRepo} {projectName}
		runCommand(exec.Command("git", "clone", "--depth=1", craftStarterRepo, projectName), false, false)
		color.Green("✓ git clone " + craftStarterRepo)

		// ddev config --project-name={projectName}
		runCommand(exec.Command("ddev", "config", "--project-name="+projectName), false, true)
		color.Green("✓ ddev config --project-name=" + projectName)
	} else {
		// git clone -b develop git@bitbucket.org:matrixcreate/{projectName}.git {projectName}
		runCommand(exec.Command("git", "clone", "-b", "develop", "git@bitbucket.org:matrixcreate/"+projectName+".git", projectName), false, false)
		color.Green("✓ git clone git@bitbucket.org:matrixcreate/" + projectName + ".git")
	}

	// ddev start
	runCommand(exec.Command("ddev", "start"), false, true)
	color.Green("✓ ddev start")

	// ddev composer install
	if fileExists("./" + projectName + "/composer.lock") {
		runCommand(exec.Command("ddev", "composer", "install"), false, true)
		color.Green("✓ ddev composer install")
	} else {
		color.Green("- No composer.lock file found. Skipping composer install")
	}

	// ddev npm install
	if fileExists("./" + projectName + "/package-lock.json") {
		runCommand(exec.Command("ddev", "npm", "install"), false, true)
		color.Green("✓ ddev npm install")
	} else {
		color.Green("- No package-lock.json file found. Skipping npm install")
	}

	// ddev craft setup/app-id --interactive=0
	runCommand(exec.Command("ddev", "craft", "setup/app-id", "--interactive=0"), false, true)
	color.Green("✓ ddev craft setup/app-id")

	// ddev craft setup/security-key
	runCommand(exec.Command("ddev", "craft", "setup/security-key"), false, true)
	color.Green("✓ ddev craft setup/security-key")

	// ddev craft setup/db --interactive=0 --driver=mysql --database=db --password=db --user=db --server=ddev-{projectName}-db --port=3306
	runCommand(exec.Command("ddev", "craft", "setup/db", "--interactive=0", "--driver=mysql", "--database=db", "--password=db", "--user=db", "--server=ddev-"+projectName+"-db", "--port=3306"), false, true)
	color.Green("✓ ddev craft setup/db")

	// ddev import-db --src=_db/db.zip
	if fileExists("./" + projectName + "/_db/db.zip") {
		runCommand(exec.Command("ddev", "import-db", "--src=_db/db.zip"), false, true)
		color.Green("✓ ddev import-db")
	} else {
		color.Green("- No _db/db.zip file found. Skipping ddev import-db")
	}

	if fresh {
		// rm -rf ./{projectName}/.git
		runCommand(exec.Command("rm", "-rf", "./"+projectName+"/.git"), false, true)
		color.Green("✓ rm -rf ./" + projectName + "/.git")

		// git init
		runCommand(exec.Command("git", "init"), false, true)
		color.Green("✓ git init")
	}

	// ddev describe
	runCommand(exec.Command("ddev", "describe"), true, true)
}

func runCommand(cmd *exec.Cmd, showOutput bool, inProject bool) {
	if inProject {
		cmd.Dir = "./" + projectName
	}

	if showOutput {
		out, err := cmd.Output()
		if err != nil {
			color.Red("Error Running: " + cmd.String())
			color.Red("Error: " + err.Error())
			os.Exit(0)
		}
		fmt.Println(string(out))
	} else {
		err := cmd.Run()
		if err != nil {
			color.Red("Error Running: " + cmd.String())
			color.Red("Error: " + err.Error())
			os.Exit(0)
		}
	}
}

func runRemoteCommand(command string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error: Unable to find home dir: %v", err)
	}
	key, err := os.ReadFile(homeDir + "/.ssh/id_rsa")
	if err != nil {
		color.Red("Error: Unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		color.Red("Error: Unable to parse private key: %v", err)
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
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
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
