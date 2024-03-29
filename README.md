# Matrix CLI #

[![Go Report Card](https://goreportcard.com/badge/github.com/MatrixCreate/matrix)](https://goreportcard.com/report/github.com/MatrixCreate/matrix)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## About ##
Matrix CLI is a command line tool for managing projects using DDEV, Github and AWS.

## Features ##
- Create a new project
- Edit an existing project
- Delete a project
- Deploy a project to AWS Lightsail
- Backup a project to AWS S3
- Self updating
- Configure Github and AWS credentials

## Requirements ##
- Go 1.16 or higher
- Github CLI
- AWS CLI
- DDEV

## Commands ##

- `matrix help` - Show help
- `matrix version` - Show version
- `matrix update` - Self update Matrix CLI
- `matrix status` - Status of Matrix CLI
- `matrix configure` - Initialize a new project
- `matrix create {name}` - Create a new project
- `matrix edit {name}` - Edit a project
- `matrix delete {name}` - Delete a project
- `matrix deploy` - Deploys the current project you are in to AWS Lightsail
- `matrix backup` - Backups the current project you are in to AWS S3
- `matrix aws --list` - List all AWS instances
- `matrix aws --spreadsheet` - Create a spreadsheet of all AWS instances
- `matrix web` - Setup web server

## Installing ##

1. [Install Go](https://go.dev/doc/install)
2. `go install github.com/MatrixCreate/matrix@latest`
3. Add to PATH...

    For ZSH add to ~/.zshrc
    ```
    # Go
    export GOROOT=/usr/local/Go
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
    ```
4. `matrix help`

## Building Executable ##

`GOOS=linux GOARCH=amd64 go build`
