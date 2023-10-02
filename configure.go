package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

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
	data := "[profile " + profileName + "]\n"
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
