package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var craftStarterRepo string = "git@github.com:MatrixCreate/craft-starter.git"
var projectName string = ""
var projectType string = "unknown"
var commandCount int = 0
var s *spinner.Spinner = spinner.New(spinner.CharSets[25], 100*time.Millisecond)

func main() {
	app := &cli.App{
		Name:      "Matrix CLI",
		Authors: []*cli.Author{
			{
				Name:  "Adam Glaysher",
				Email: "adam@matrixcreate.com",
			},
			{
				Name:  "Jamie Adams",
				Email: "jamie@matrixcreate.com",
			},
		},
		Version:   "v2.0.0",
		Copyright: "(c) 2023 Matrix Create",
		Usage:     "Project Management CLI Tool",
		Commands: []*cli.Command{
			{
				Name:   "status",
				Aliases: []string{"s"},
				Usage:  "Show status of Matrix CLI",
				Action: func(cCtx *cli.Context) error {
					color.Magenta("Matrix CLI Status")

					// Check if Git is installed
					runCommand(exec.Command("git", "--version"), true, false, false)

					color.Green("✓ Git is installed")

					// Check if DDEV is installed
					runCommand(exec.Command("ddev", "--version"), true, false, false)

					color.Green("✓ DDEV is installed")

					// Check if Github CLI (gh) is installed and authed
					runCommand(exec.Command("gh", "auth", "status"), true, false, false)

					color.Green("✓ GitHub CLI installed and authed")

					// Check if AWS CLI (aws) is installed and authed
					runCommand(exec.Command("aws", "--version"), true, false, false)

					color.Green("✓ AWS CLI installed and authed")

					color.Green("✓ Completed: Matrix CLI Status")

					return nil
				},
			},
			{
				Name:   "configure",
				Aliases: []string{"c", "config"},
				Usage:  "Configure Matrix CLI with AWS IAM Identity Center and Github CLI",
				Action: func(cCtx *cli.Context) error {
					configureMatrix()
					configureAWS()
					configureGithub()

					return nil
				},
			},
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
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy project to AWS Lightsail",
				Action: func(cCtx *cli.Context) error {
					// Deploy a lightsail instance using the git repo from the current directory
					color.Magenta("Deploying project to AWS Lightsail")
					
					var blueprintID string = "lamp_8_bitnami"
					var profileName string = "matrix"

					// Get project name
					projectName = cCtx.Args().First()

					if projectName == "" {
						color.Red("× Error: Missing project name")
						os.Exit(1)
					}

					// Check if project is craft
					if fileExists("craft") {
						projectType = "craft"
					}

					// Check if project is wordpress
					if fileExists("wp-content") {
						projectType = "wordpress"
					}

					// Run gh auth token to get github token and store in variable
					cmd := exec.Command("gh", "auth", "token")
					out, err := cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}
					githubToken := string(out)

					// Get current git remote url
					cmd = exec.Command("git", "config", "--get", "remote.origin.url")
					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}
					gitRemoteUrl := string(out)

					// Convert git remote URL to HTTPS
					if string(out[0:3]) == "git" {
						gitRemoteUrl = "https://" + string(out[4:len(out)-5])

						// github.com: should be github.com/
						gitRemoteUrl = strings.Replace(gitRemoteUrl, "github.com:", "github.com/", 1)
					}

					// Get current github username
					cmd = exec.Command("gh", "api", "user")
					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}
					githubUsernameJson := string(out)

					// Get github username from JSON
					githubUsername := strings.Split(githubUsernameJson, "\"")[3]

					color.White("Git Username: " + githubUsername)
					color.White("Git Remote URL: " + gitRemoteUrl)
					color.White("Git Token: " + githubToken)

					// Add username:token to git remote URL
					gitRemoteUrl = strings.Replace(gitRemoteUrl, "https://", "https://"+githubUsername+":"+githubToken+"@", 1)

					// Remove new line
					gitRemoteUrl = strings.Replace(gitRemoteUrl, "\n", "", 1)

					color.White("Git Remote URL: " + gitRemoteUrl)

					// Start creating deploy script
					data :=	"#!/bin/bash\n"
					
					// cd to htdocs directory
					data += "cd /home/bitnami/htdocs\n"

					// Remove index.html
					data += "rm index.html\n"

					if projectType == "craft" {
						// Edit /opt/bitnami/apache/conf/bitnami/bitnami.conf and change /home/bitnami/htdocs to /home/bitnami/htdocs/web and save
						data += "sed -i 's/\\/home\\/bitnami\\/htdocs/\\/home\\/bitnami\\/htdocs\\/web/g' /opt/bitnami/apache/conf/bitnami/bitnami.conf\n"
					
						// Restart Apache
						data += "sudo /opt/bitnami/ctlscript.sh restart apache\n"
					}

					// git clone repo into current directory
					data += "git clone " + gitRemoteUrl + " .\n"

					// chown -R bitnami:daemon /home/bitnami/htdocs
					data += "chown -R bitnami:daemon /home/bitnami/htdocs/\n"

					// if composer file exists on the instance then run composer install
					data += "if [ -f \"composer.json\" ]; then\n"
					data += "composer install --no-dev --no-interaction --prefer-dist --optimize-autoloader\n"
					data += "fi\n"

					// if package.json file exists on the instance then run npm install
					data += "if [ -f \"package.json\" ]; then\n"
					data += "npm install\n"
					data += "fi\n"

					// if craft file exists on the instance then run craft setup commands
					if projectType == "craft" {
						// php craft setup/app-id --interactive=0
						data += "php craft setup/app-id --interactive=0\n"

						// php craft setup/security-key
						data += "php craft setup/security-key\n"
					}
					
					// TODO: Database setup and migrations etc...

					// Install Go
					data += "sudo apt -y install golang-go\n"

					// Install Matrix CLI
					data += "go install github.com/MatrixCreate/matrix@latest\n"
					
					// If deploy.sh script already exists then make a copy and delete it
					if fileExists("deploy.sh") {
						runCommand(exec.Command("mv", "deploy.sh", "deploy.sh.old"), false, false, true)
					}

					// Write to deploy.sh
					color.White("Writing to: deploy.sh")
					f, err := os.OpenFile("deploy.sh", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log.Fatal(err)
					}
					defer f.Close()
					if _, err := f.WriteString(data); err != nil {
						log.Fatal(err)
					}
					
					color.Green("✓ Completed: Writing to: deploy.sh")

					// Deploy a lightsail instance using the git repo from the current directory
					cmd = exec.Command("aws", "lightsail", "create-instances", "--instance-names", projectName, "--availability-zone", "eu-west-2a", "--blueprint-id", blueprintID, "--bundle-id", "nano_3_0", "--user-data", "file://deploy.sh", "--profile", profileName)

					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}
					fmt.Println(string(out))

					// Delete deploy.sh file
					runCommand(exec.Command("rm", "deploy.sh"), false, false, true)

					color.Green("✓ Completed: Deploying project to AWS Lightsail")

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            DEPLOYMENT STARTED                🎉")
					color.Magenta("--------------------------------------------------")

					// Check if instance is running and wait until it is
					color.Magenta("Checking if instance is running")

					cmd = exec.Command("aws", "lightsail", "get-instance-state", "--instance-name", projectName, "--profile", profileName)

					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}
					fmt.Println(string(out))

					// Wait until instance is running
					for strings.Contains(string(out), "pending") {
						color.Magenta("Waiting for instance to start")

						cmd = exec.Command("aws", "lightsail", "get-instance-state", "--instance-name", projectName, "--profile", profileName)

						out, err = cmd.Output()
						if err != nil {
							color.Red("× Error Running: " + cmd.String())
							color.Red("× " + err.Error())
							os.Exit(1)
						}
						fmt.Println(string(out))

						time.Sleep(5 * time.Second)
					}

					color.Green("✓ Completed: Instance is running")

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            DEPLOYMENT COMPLETE               🎉")
					color.Magenta("--------------------------------------------------")

					// Get IP of the new instance as a var
					cmd = exec.Command("aws", "lightsail", "get-instance", "--instance-name", projectName, "--profile", profileName)

					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}

					// Get ['instance']['publicIpAddress'] from JSON and not anything after
					instancePublicIpAddress := strings.Split(string(out), "\"publicIpAddress\": \"")[1]
					instancePublicIpAddress = strings.Split(instancePublicIpAddress, "\"")[0]

					// Print IP
					color.White("http://" + instancePublicIpAddress)

					return nil
				},
			},
			{
				Name:    "backup",
				Aliases: []string{"b"},
				Usage:   "Backup project to S3",
				Action: func(cCtx *cli.Context) error {
					color.Magenta("Backing up project to AWS S3")

					projectName = cCtx.Args().First()

					if projectName == "" {
						color.Red("× Error: Missing project name")
						os.Exit(1)
					}

					// Check if project is craft
					if fileExists("/craft") {
						projectType = "craft"
					}

					// Check if project is wordpress
					if fileExists("/wp-content") {
						projectType = "wordpress"
					}

					// Mysql server info
					var mysqlHost string = projectName
					var mysqlPort string = "3306"
					var mysqlUser string = "db"
					var mysqlPass string = "db"
					var mysqlDatabase string = "db"

					// Get mysql server info from .env
					if projectType == "craft" && fileExists(projectName + "/.env") {
						err := godotenv.Load(projectName + "/.env")
						if err != nil {
							log.Fatal("Error loading .env file")
						}

						mysqlHost = os.Getenv("DB_HOST")
						mysqlPort = os.Getenv("DB_PORT")
						mysqlUser = os.Getenv("DB_USER")
						mysqlPass = os.Getenv("DB_PASSWORD")
						mysqlDatabase = os.Getenv("DB_DATABASE")
					}

					// Get mysql server info from wp-config.php
					if projectType == "wordpress" && fileExists("/wp-config.php") {
						// TODO: Get mysql server info from wp-config.php
					}

					// Use these details to create a sql backup file using mysqldump
					runCommand(exec.Command("mysqldump", "-h", mysqlHost, "-P", mysqlPort, "-u", mysqlUser, "-p"+mysqlPass, mysqlDatabase, "--single-transaction", "--quick", "--lock-tables=false", "--result-file="+projectName+".sql"), false, true, false)

					// Zip up the sql backup file
					runCommand(exec.Command("zip", "-r", projectName+".zip", projectName+".sql"), false, true, false)

					// Change file name to include date
					runCommand(exec.Command("mv", projectName+".zip", projectName+"-"+time.Now().Format("2006-01-02")+".zip"), false, true, false)

					// Upload to AWS S3
					runCommand(exec.Command("aws", "s3", "cp", projectName+"-"+time.Now().Format("2006-01-02")+".zip", "s3://matrixcreate/"+projectName+"/database", "--profile", "matrix"), false, true, false)

					// Create zip of project files
					runCommand(exec.Command("zip", "-r", projectName+".zip"), false, true, false)

					// Change file name to include date
					runCommand(exec.Command("mv", projectName+".zip", projectName+"-"+time.Now().Format("2006-01-02")+".zip"), false, true, false)

					// Upload to AWS S3
					runCommand(exec.Command("aws", "s3", "cp", projectName+"-"+time.Now().Format("2006-01-02")+".zip", "s3://matrixcreate/"+projectName+"/files", "--profile", "matrix"), false, true, false)
				
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

func configureMatrix() {
	// If config already setup then exit
	if fileExists(os.Getenv("HOME") + "/.matrix/config") {
		color.Green("✓ Matrix CLI Already Configured")

		return
	}

	// Create config file
	userMatrixPath := os.Getenv("HOME") + "/.matrix"
	userMatrixConfigPath := userMatrixPath + "/config"

	// AWS SSO config vars
	var awsRegion string = ""
	var awsAccountID string = ""
	var awsStartURL string = ""
	var awsRoleName string = ""

	// Promp for input for each AWS SSO config vars
	color.White("Enter AWS Region")
	fmt.Scan(&awsRegion)
	color.White("Enter AWS Account ID")
	fmt.Scan(&awsAccountID)
	color.White("Enter AWS Start URL")
	fmt.Scan(&awsStartURL)
	color.White("Enter AWS Role Name")
	fmt.Scan(&awsRoleName)

	// Setup aws config data
	userMatrixConfigData := "aws_region = " + awsRegion + "\n"
	userMatrixConfigData += "aws_account_id = " + awsAccountID + "\n"
	userMatrixConfigData += "aws_start_url = " + awsStartURL + "\n"
	userMatrixConfigData += "aws_role_name = " + awsRoleName + "\n\n"

	if !fileExists(userMatrixPath) {
		runCommand(exec.Command("mkdir", "-p", userMatrixPath), false, false, true)
	}

	// Write to ~/.matrix/config
	color.White("Writing to: " + userMatrixConfigPath)
	f, err := os.OpenFile(userMatrixConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(userMatrixConfigData); err != nil {
		log.Fatal(err)
	}

	color.Green("✓ Completed: Writing to: " + userMatrixConfigPath)

	color.Magenta("--------------------------------------------------")
	color.Magenta("🎉            MATRIX POWER ACTIVATED            🎉")
	color.Magenta("--------------------------------------------------")
}

func configureGithub() {
	// run command: gh auth login and allow to reply to prompt
	s.Stop()
	cmd := exec.Command("gh", "auth", "login", "--git-protocol", "ssh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	color.Magenta("--------------------------------------------------")
	color.Magenta("🎉            GITHUB POWER ACTIVATED            🎉")
	color.Magenta("--------------------------------------------------")
}

func configureAWS() {
	color.Magenta("Configuring Matrix CLI with AWS IAM Identity Center")

	userAwsPath := os.Getenv("HOME") + "/.aws"
	profileName := "matrix"

	// Get AWS SSO config vars from ~/.matrix/config
	var ssoRegion string = ""
	var ssoAccountID string = ""
	var ssoStartUrl string = ""
	var ssoRoleName string = ""

	if fileExists(os.Getenv("HOME") + "/.matrix/config") {
		err := godotenv.Load(os.Getenv("HOME") + "/.matrix/config")
		if err != nil {
			log.Fatal("Error loading .matrix/config file")
		}
		
		ssoRegion = os.Getenv("aws_region")
		ssoAccountID = os.Getenv("aws_account_id")
		ssoStartUrl = os.Getenv("aws_start_url")
		ssoRoleName = os.Getenv("aws_role_name")
	} else {
		color.Red("× Error: Missing ~/.matrix/config file")
		os.Exit(1)
	}

	if !fileExists(userAwsPath + "/config") {
		runCommand(exec.Command("mkdir", "-p", userAwsPath), false, false, true)
	} else {
		// Rename old config file
		runCommand(exec.Command("mv", userAwsPath+"/config", userAwsPath+"/config.old"), false, false, true)
	}

	// Setup aws config data
	data :=	"[profile " + profileName + "]\n"
	data += "sso_session = matrix-sso\n"
	data += "sso_role_name = " + ssoRoleName + "\n"
	data += "sso_account_id = " + ssoAccountID + "\n"
	data += "region = " + ssoRegion + "\n"
	data += "output = json\n\n"
	data += "[sso-session matrix-sso]\n"
	data += "sso_region = " + ssoRegion + "\n"
	data += "sso_start_url = " + ssoStartUrl + "\n"
	data += "sso_registration_scopes = sso:account:access\n"

	// Write to ~/.aws/config
	color.White("Writing to: " + userAwsPath + "/config")
	f, err := os.OpenFile(userAwsPath+"/config", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		log.Fatal(err)
	}
	color.Green("✓ Completed: Writing to: " + userAwsPath + "/config")

	// Redirect to login to AWS SSO
	cmd := exec.Command("aws", "sso", "login", "--profile", profileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	color.Magenta("--------------------------------------------------")
	color.Magenta("🎉            AWS POWER ACTIVATED               🎉")
	color.Magenta("--------------------------------------------------")
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
				color.White("Tip: Run the above command seperately for more info to find out what went wrong")
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
