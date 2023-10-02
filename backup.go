package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func backup(cCtx *cli.Context) {
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
}
