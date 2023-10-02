package main

import (
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
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
					status()

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
					create(cCtx)

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
					edit(cCtx)

					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "Stop and delete project",
				Action: func(cCtx *cli.Context) error {
					delete(cCtx)

					return nil
				},
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy project to AWS Lightsail",
				Action: func(cCtx *cli.Context) error {
					deploy(cCtx)

					return nil
				},
			},
			{
				Name:    "backup",
				Aliases: []string{"b"},
				Usage:   "Backup project to S3",
				Action: func(cCtx *cli.Context) error {
					backup(cCtx)

					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"self-update"},
				Usage:   "Self Update Matrix CLI",
				Action: func(cCtx *cli.Context) error {
					update()

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
