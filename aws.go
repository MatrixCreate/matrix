package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"github.com/xuri/excelize/v2"
)

func listInstances() {
	var lightsailInstancesJSON map[string]interface{} = getLightsailInstancesAsJSON()
	var ec2InstancesJSON map[string]interface{} = getEC2InstancesAsJSON()

	color.Magenta("EC2 Instances:")

	// Loop through ec2 instances
	for _, reservation := range ec2InstancesJSON["Reservations"].([]interface{}) {
		reservation := reservation.(map[string]interface{})
		for _, instance := range reservation["Instances"].([]interface{}) {
			instance := instance.(map[string]interface{})
			color.White("  - " + instance["InstanceId"].(string))

			if instance["State"].(map[string]interface{})["Name"].(string) == "running" {
				color.Green("    - Status: " + instance["State"].(map[string]interface{})["Name"].(string))
			} else {
				color.Red("    - Status: " + instance["State"].(map[string]interface{})["Name"].(string))
			}

			color.White("    - Public IP: " + instance["PublicIpAddress"].(string))
			color.White("    - Private IP: " + instance["PrivateIpAddress"].(string))
		}
	}

	color.Magenta("Lightsail Instances:")

	// Loop through lightsail instances
	for _, instance := range lightsailInstancesJSON["instances"].([]interface{}) {
		instance := instance.(map[string]interface{})
		color.White("  - " + instance["name"].(string))

		if instance["state"].(map[string]interface{})["name"].(string) == "running" {
			color.Green("    - Status: " + instance["state"].(map[string]interface{})["name"].(string))
		} else {
			color.Red("    - Status: " + instance["state"].(map[string]interface{})["name"].(string))
		}

		color.White("    - Public IP: " + instance["publicIpAddress"].(string))
		color.White("    - Private IP: " + instance["privateIpAddress"].(string))
	}
}

func createSpreadsheetOfInstances(cCtx *cli.Context) {
	s.Suffix = " Creating spreadsheet of AWS instances..."
	s.Start()

	var lightsailInstancesJSON map[string]interface{} = getLightsailInstancesAsJSON()
	var ec2InstancesJSON map[string]interface{} = getEC2InstancesAsJSON()

	// Create a new spreadsheet
	f := excelize.NewFile()

	// Create a new sheet
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		color.Red("× Error Creating Spreadsheet: " + err.Error())
		os.Exit(1)
	}

	// Add data
	f.SetCellValue("Sheet1", "A1", "EC2 or Lightsail")
	f.SetCellValue("Sheet1", "B1", "Name")
	f.SetCellValue("Sheet1", "C1", "Status")
	f.SetCellValue("Sheet1", "D1", "Instance Type")
	f.SetCellValue("Sheet1", "E1", "Public IP")
	f.SetCellValue("Sheet1", "F1", "Private IP")
	f.SetCellValue("Sheet1", "G1", "CPUs")
	f.SetCellValue("Sheet1", "H1", "RAM")
	f.SetCellValue("Sheet1", "I1", "Disk")
	f.SetCellValue("Sheet1", "J1", "Region")

	// Loop through lightsail instances
	for i, instance := range lightsailInstancesJSON["instances"].([]interface{}) {
		instance := instance.(map[string]interface{})
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), "Lightsail")
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), instance["name"].(string))
		f.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), instance["state"].(map[string]interface{})["name"].(string))
		f.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), instance["bundleId"].(string))
		f.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), instance["publicIpAddress"].(string))
		f.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), instance["privateIpAddress"].(string))
		f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), instance["hardware"].(map[string]interface{})["cpuCount"].(float64))
		f.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), instance["hardware"].(map[string]interface{})["ramSizeInGb"].(float64))

		// Disk size
		var diskSize float64
		for _, disk := range instance["hardware"].(map[string]interface{})["disks"].([]interface{}) {
			disk := disk.(map[string]interface{})
			diskSize += disk["sizeInGb"].(float64)
		}
		f.SetCellValue("Sheet1", "I"+strconv.Itoa(i+2), strconv.FormatFloat(diskSize, 'f', 0, 64)+" GB")

		f.SetCellValue("Sheet1", "J"+strconv.Itoa(i+2), instance["location"].(map[string]interface{})["availabilityZone"].(string))
	}

	// Loop through ec2 instances and add on to the end of the lightsail instances
	for i, reservation := range ec2InstancesJSON["Reservations"].([]interface{}) {
		reservation := reservation.(map[string]interface{})
		for _, instance := range reservation["Instances"].([]interface{}) {
			instance := instance.(map[string]interface{})
			f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), "EC2")
			f.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), instance["InstanceId"].(string))
			f.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), instance["State"].(map[string]interface{})["Name"].(string))
			f.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), instance["InstanceType"].(string))
			f.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), instance["PublicIpAddress"].(string))
			f.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), instance["PrivateIpAddress"].(string))
			f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), instance["CpuOptions"].(map[string]interface{})["CoreCount"].(float64))
		}
	}

	// Set active sheet of the workbook
	f.SetActiveSheet(index)

	// Set column widths
	f.SetColWidth("Sheet1", "A", "A", 15)
	f.SetColWidth("Sheet1", "B", "B", 20)
	f.SetColWidth("Sheet1", "C", "C", 15)
	f.SetColWidth("Sheet1", "D", "D", 15)
	f.SetColWidth("Sheet1", "E", "E", 15)
	f.SetColWidth("Sheet1", "F", "F", 15)
	f.SetColWidth("Sheet1", "G", "G", 10)
	f.SetColWidth("Sheet1", "H", "H", 10)
	f.SetColWidth("Sheet1", "I", "I", 10)
	f.SetColWidth("Sheet1", "J", "J", 15)

	// Set row heights
	for i := 1; i <= len(ec2InstancesJSON["Reservations"].([]interface{}))+len(lightsailInstancesJSON["instances"].([]interface{}))+1; i++ {
		f.SetRowHeight("Sheet1", i, 20)
	}

	// Save spreadsheet
	if err := f.SaveAs("aws-instances.xlsx"); err != nil {
		color.Red("× Error Saving Spreadsheet: " + err.Error())
		os.Exit(1)
	}

	color.Magenta("--------------------------------------------------")
	color.Magenta("🎉   SPREADSHEET CREATED: aws-instances.xlsx    🎉")
	color.Magenta("--------------------------------------------------")
}

func getLightsailInstancesAsJSON() map[string]interface{} {
	// Run aws lightsail get-instances
	cmd := exec.Command("aws", "lightsail", "get-instances", "--profile", "matrix")
	out, err := cmd.Output()
	if err != nil {
		color.Red("× Error Running: " + cmd.String())
		color.Red("× " + err.Error())
		os.Exit(1)
	}
	lightsailInstances := string(out)

	// Convert output to json
	var lightsailInstancesJSON map[string]interface{}
	err = json.Unmarshal([]byte(lightsailInstances), &lightsailInstancesJSON)
	if err != nil {
		color.Red("× Error Parsing JSON: " + err.Error())
		os.Exit(1)
	}

	s.Stop()

	return lightsailInstancesJSON
}

func getEC2InstancesAsJSON() map[string]interface{} {
	// Run aws ec2 describe-instances
	cmd := exec.Command("aws", "ec2", "describe-instances", "--profile", "matrix")
	out, err := cmd.Output()
	if err != nil {
		color.Red("× Error Running: " + cmd.String())
		color.Red("× " + err.Error())
		os.Exit(1)
	}
	ec2Instances := string(out)

	// Convert output to json
	var ec2InstancesJSON map[string]interface{}
	err = json.Unmarshal([]byte(ec2Instances), &ec2InstancesJSON)
	if err != nil {
		color.Red("× Error Parsing JSON: " + err.Error())
		os.Exit(1)
	}

	return ec2InstancesJSON
}
