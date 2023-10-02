package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var craftStarterRepo string = "git@github.com:MatrixCreate/craft-starter.git"
var projectName string = ""
var projectType string = ""
var commandCount int = 0
var s *spinner.Spinner = spinner.New(spinner.CharSets[25], 100*time.Millisecond)

func main() {
	app := &cli.App{
		Name: "Matrix CLI",
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
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Show status of Matrix CLI",
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

					color.Green("✓ GitHub CLI is installed and authed")

					// TODO: Check if AWS CLI (aws) is installed and authed
					runCommand(exec.Command("aws", "--version"), true, false, false)

					color.Green("✓ AWS CLI is installed and authed")

					color.Green("✓ Completed: Matrix CLI Status")

					return nil
				},
			},
			{
				Name:    "configure",
				Aliases: []string{"c", "config"},
				Usage:   "Configure Matrix CLI with AWS IAM Identity Center and Github CLI",
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
					// Deploy an AWS EC2 instance using the Git repo from the current directory using launch template
					color.Magenta("Deploying project to AWS")

					var launchTemplateName string = "matrix-2023-10-01"
					var instanceType string = "t2.micro"
					var profileName string = "matrix"

					// Get project name
					projectName = cCtx.Args().First()

					if projectName == "" {
						// Get project name from current directory
						workingDir, err := os.Getwd()
						if err != nil {
							color.Red("× Error: " + err.Error())
							os.Exit(1)
						}

						projectName = strings.Split(workingDir, "/")[len(strings.Split(workingDir, "/"))-1]

						color.White("Project Name: " + projectName)
					}

					// Check if project is craft
					if fileExists("craft") {
						projectType = "craft"
					}

					// Check if project is wordpress
					if fileExists("wp-content") {
						projectType = "wordpress"
					}

					// Get github token
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
					data := "#!/bin/bash\n"

					// cd to /var/www/html
					data += "cd /var/www/html\n"

					// if projectType == "craft" {
					// 	// Edit /opt/bitnami/apache/conf/bitnami/bitnami.conf and change /opt/bitnami/apache/htdocs to /opt/bitnami/apache/htdocs/web and save
					// 	data += "sed -i 's|/opt/bitnami/apache/htdocs|/opt/bitnami/apache/htdocs/web|g' /opt/bitnami/apache2/conf/bitnami/bitnami.conf\n"

					// 	// Restart Apache
					// 	data += "/opt/bitnami/ctlscript.sh restart apache\n"
					// }

					// git clone repo into current directory
					data += "git clone " + gitRemoteUrl + " .\n"

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

					// Run a EC2 instance using the git repo from the current directory
					cmd = exec.Command("aws", "ec2", "run-instances", "--launch-template", "LaunchTemplateName="+launchTemplateName, "--instance-type", instanceType, "--user-data", "file://deploy.sh", "--tag-specifications", "ResourceType=instance,Tags=[{Key=Name,Value="+projectName+"}]", "--profile", profileName)
					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						// More error details (TODO: Add to more errors)
						color.Red("× " + string(err.(*exec.ExitError).Stderr))
						os.Exit(1)
					}

					// Get instance ID from JSON
					instanceID := strings.Split(string(out), "\"InstanceId\": \"")[1]
					instanceID = strings.Split(instanceID, "\"")[0]

					color.White("Instance ID: " + instanceID)

					// Delete deploy.sh file
					runCommand(exec.Command("rm", "deploy.sh"), false, false, true)

					color.Green("✓ Completed: Started new EC2 instance")

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            DEPLOYMENT STARTED                🎉")
					color.Magenta("--------------------------------------------------")

					// Check if instance is running and wait until it is
					color.Magenta("Checking if instance is running")

					var instanceState string = "pending"

					// Set spinner message
					s.Suffix = " Instance Status: " + instanceState

					// Start the spinner
					s.Start()

					for instanceState != "running" {
						cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-ids", instanceID, "--profile", profileName)

						out, err = cmd.Output()
						if err != nil {
							color.Red("× Error Running: " + cmd.String())
							color.Red("× " + err.Error())
							os.Exit(1)
						}

						// Get ['Reservations'][0]['Instances'][0]['State']['Name']
						var result map[string]interface{}
						json.Unmarshal([]byte(out), &result)
						instanceState = result["Reservations"].([]interface{})[0].(map[string]interface{})["Instances"].([]interface{})[0].(map[string]interface{})["State"].(map[string]interface{})["Name"].(string)

						// Update spinner status
						s.Suffix = " Instance Status: " + instanceState

						// Wait 10 seconds
						time.Sleep(10 * time.Second)
					}

					// Stop the spinner
					s.Stop()

					color.Green("✓ Success: Instance is running")

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            DEPLOYMENT COMPLETE               🎉")
					color.Magenta("--------------------------------------------------")

					// Get IP of the new instance as a var
					cmd = exec.Command("aws", "ec2", "describe-instances", "--filters", "Name=tag:Name,Values="+projectName, "--profile", profileName)

					out, err = cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}

					// Get ['Reservations'][0]['Instances'][0]['PublicIpAddress']
					instancePublicIpAddress := strings.Split(string(out), "\"PublicIpAddress\": \"")[1]
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

					// Check if mysql is installed or exit
					cmd := exec.Command("mysql", "--version")
					_, err := cmd.Output()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.Red("× " + err.Error())
						os.Exit(1)
					}

					color.Green("✓ MySQL is installed")

					// Get project name
					projectName = cCtx.Args().First()
					if projectName == "" {
						color.Red("× Error: Missing project name")
						os.Exit(1)
					}

					// Check if project is craft
					if fileExists("./craft") {
						projectType = "craft"

						color.Green("✓ Craft Detected")

						// check if .env file
						if !fileExists("./.env") {
							color.Red("× Error: Missing .env file")
							os.Exit(1)
						}

						// Get DB settings from .env file
						err := godotenv.Load("./.env")
						if err != nil {
							log.Fatal("Error loading .env file")
						}

						// Get DB settings from .env file
						var dbDriver string = os.Getenv("DB_DRIVER")
						var dbServer string = os.Getenv("DB_SERVER")
						var dbPort string = os.Getenv("DB_PORT")
						var dbUser string = os.Getenv("DB_USER")
						var dbPassword string = os.Getenv("DB_PASSWORD")
						var dbName string = os.Getenv("DB_DATABASE")

						// Check if DB settings are empty
						if dbDriver == "" || dbServer == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
							color.Red("× Error: Missing DB settings in .env file")
							os.Exit(1)
						}

						// backup the database using mysqldump
						cmd = exec.Command(
							"mysqldump",
							"-u", dbUser,
							"-p"+dbPassword,
							"-h", dbServer,
							"-P", dbPort,
							dbName,
							"--single-transaction",
							"--quick",
							"--lock-tables=false",
							"--routines",
							"--triggers",
							"--events",
							"--skip-comments",
							"--skip-dump-date",
							"--skip-set-charset",
							"--skip-add-locks",
							"--skip-disable-keys",
							"--skip-tz-utc",
							"--skip-lock-tables",
							"--result-file="+projectName+".sql",
						)

						color.White("Running: " + cmd.String())

						err = cmd.Run()
						if err != nil {
							color.Red("× Error Running: " + cmd.String())
							color.Red("× " + err.Error())
							os.Exit(1)
						}

						color.Green("✓ Completed: Database backup file created locally")
					}

					// Check if project is wordpress
					if fileExists("./wp-content") {
						projectType = "wordpress"

						color.White("✓ WordPress Detected")

						// Exit as currently don't support wordpress backups
						color.Red("× Error: WordPress backups are not currently supported")
						os.Exit(1)
					}

					var backupFileName = projectName + "-" + time.Now().Format("2006-01-02-15-04-05")

					runCommand(exec.Command("tar", "-czf", backupFileName+".sql.tar.gz", projectName+".sql"), false, false, false)
					runCommand(exec.Command("tar", "--warning=no-file-changed", "-czf", backupFileName+".tar.gz", "--exclude='web/cpresources'", "--exclude='storage/runtime'", "--exclude='vendor'", "--exclude='.git'", "--exclude='"+backupFileName+".tar.gz'", "."), false, false, false)

					// Upload SQL backup to S3
					cmd = exec.Command("aws", "s3", "cp", backupFileName+".sql.tar.gz", "s3://"+projectName+"/backups/"+backupFileName+".sql.tar.gz")
					color.White("Running: " + cmd.String())
					err = cmd.Run()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.White("Your AWS token probably has expired. Run 'matrix configure' to setup AWS CLI Auth again")
						os.Exit(1)
					}

					// Upload Files backup to S3
					cmd = exec.Command("aws", "s3", "cp", backupFileName+".tar.gz", "s3://"+projectName+"/backups/"+backupFileName+".tar.gz")
					color.White("Running: " + cmd.String())
					err = cmd.Run()
					if err != nil {
						color.Red("× Error Running: " + cmd.String())
						color.White("Your AWS token probably has expired. Run 'matrix configure' to setup AWS CLI Auth again")
						os.Exit(1)
					}

					color.Green("✓ Completed: Database backup uploaded to S3")

					// Delete local temp files
					runCommand(exec.Command("rm", backupFileName+".sql.tar.gz"), false, false, true)
					runCommand(exec.Command("rm", backupFileName+".tar.gz"), false, false, true)
					runCommand(exec.Command("rm", projectName+".sql"), false, false, true)

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            BACKUP COMPLETE                   🎉")
					color.Magenta("--------------------------------------------------")

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

					color.Magenta("--------------------------------------------------")
					color.Magenta("🎉            UPDATE COMPLETE                   🎉")
					color.Magenta("--------------------------------------------------")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
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
