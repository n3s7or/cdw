package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/n3s7or/cdw/pkg/commands"
	"github.com/n3s7or/cdw/pkg/common"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name:                   "cdw",
		Usage:                  "https://github.com/n3s7or/cdw#readme",
        Commands:               []*cli.Command{&commands.ProjectsCommand, &commands.BuildsCommand},
        EnableBashCompletion:   true,
        Action: func(c *cli.Context) error {

            cfg, err := config.LoadDefaultConfig(c.Context)
            if err != nil {
                log.Fatal(err.Error())
            	return err
            }

            projects, err := commands.ListAllProjects(c, &cfg)
            if err != nil {
                log.Fatal(err.Error())
                return err
            }
            
            in := survey.Select{Message: "Select project:", Options: projects}
            var selectedIndex int
            err = survey.AskOne(&in, &selectedIndex)
            if err != nil {
                if err == terminal.InterruptErr {
                    fmt.Println("Interrupted by user, exiting...")
                    return nil
                }
                log.Fatal(err.Error())
                return err
            }

            buildsListNames, err := naws.ListBuilds(c, &cfg, &projects[selectedIndex])
            if err != nil {
                println(err.Error())
                return nil
            }

            if len(buildsListNames) == 0 {
                fmt.Printf("No builds available for project `%s`\n", projects[selectedIndex])
                return nil
            }

            builds, err := common.GetBuildsFromBuildsName(c, cfg, buildsListNames)
            if err != nil {
                log.Fatal(err.Error())
            }
            
            surveyBuildsOptions := common.GenerateOptionsListForsurvey(builds)
            in = survey.Select{Message: "Select build:", Options: surveyBuildsOptions}
            err = survey.AskOne(&in, &selectedIndex)
            if err != nil {
                if err == terminal.InterruptErr {
                    fmt.Println("Interrupted by user, exiting...")
                    return nil
                }
                log.Fatal(err.Error())
                return err
            }

            common.MonitorBuild(c, &cfg, &builds[selectedIndex], "")

            return nil
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}
