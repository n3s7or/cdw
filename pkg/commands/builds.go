package commands

import (
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/n3s7or/cdw/pkg/common"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

var BuildsCommand = cli.Command{
	Name: "builds",
	Usage: "List builds for a provided project",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "project", Required: true, Usage: "project name to check builds"},
		&cli.BoolFlag{Name: "prompt", Usage: "If provided, prompts user to select one build among the last 10 builds, show last build logs otherwise"},
		&cli.BoolFlag{Name: "no-logs", Usage: "If provided, only checks build state and transition to complete"},
	},
	Action: func(ctx *cli.Context) error {

		prompt := ctx.Bool("prompt")
		showLogs := !ctx.Bool("no-logs")

		project := ctx.String("project")
		if project == ""{
			log.Fatal("no project provided")
		}

		cfg, err := config.LoadDefaultConfig(ctx.Context)
		if err != nil {
			log.Fatal(err.Error())
			return err
		}		
		
		buildsName, err := naws.ListBuilds(ctx, &cfg, &project)
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(buildsName) == 0 {
			fmt.Printf("No builds available for project `%s`\n", project)
			return nil
		}

		builds, err := common.GetBuildsFromBuildsName(ctx, cfg, buildsName)
		if err != nil {
			log.Fatal(err.Error())
		}

		selectedIndex := 0
		if prompt {
			surveyOptions := common.GenerateOptionsListForsurvey(builds)
			in := survey.Select{Message: "Select build:", Options: surveyOptions}
			
			err = survey.AskOne(&in, &selectedIndex)
			if err != nil {
				if err == terminal.InterruptErr {
					fmt.Println("Interrupted by user, exiting...")
					return nil
				}
				log.Fatal(err.Error())
			}
		}
		
		if showLogs{
			err = common.MonitorBuild(ctx, &cfg, &builds[selectedIndex])
			if err != nil {
				log.Fatal(err.Error())
			}
			return nil
		}

		err = common.MonitorBuildState(ctx, &cfg, &builds[selectedIndex])
		if err != nil {
			log.Fatal(err.Error())
		}
		return nil
	},
}
