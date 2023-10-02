package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func deploy(cCtx *cli.Context) {
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
}
