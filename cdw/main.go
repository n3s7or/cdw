package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go-v2/config"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/gen2brain/beeep"
	"github.com/n3s7or/cdw/pkg/naws"
    "github.com/n3s7or/cdw/pkg/commands"
	"github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name:                 "cdw",
		Usage:                "https://github.com/n3s7or/cdw#readme",
        Commands:   []*cli.Command{&commands.ProjectsCommand}, // todo: add commands
        Action: func(c *cli.Context) error {

            cfg, err := config.LoadDefaultConfig(c.Context)
            if err != nil {
                log.Fatal(err.Error())
            	return err
            }

            cbClient := naws.GetClient(c, &cfg)

            var cbNextToken *string // ToDo: get other projects if there is a next token
            projects, err := naws.ListProjects(c, cbClient, cbNextToken)
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

            buildsListNames, err := naws.ListBuilds(c, cbClient, &projects[selectedIndex])
            if err != nil {
                println(err.Error())
                return nil
            }

            if len(buildsListNames) == 0 {
                fmt.Printf("No builds available for project `%s`\n", projects[selectedIndex])
                return nil
            }

            var builds []cbtypes.Build
            if len(buildsListNames) > 9 {
                builds, err = naws.GetBuildsInfo(c, cbClient, buildsListNames[0:10])
            }else {
                builds, err = naws.GetBuildsInfo(c, cbClient, buildsListNames)
            }
            if err != nil {
                log.Fatal(err.Error())
                return nil
            }
            
            var listBuilds []string;
            for _, item := range(builds){
                id := string(strings.Split((*item.Id), ":")[1])
                initiator := string(*item.Initiator)
                opt := fmt.Sprintf("%-*s\t%s\t(%s)", 11, fmt.Sprintf("[%s]", string(item.BuildStatus)), initiator, id)
                listBuilds = append(listBuilds, opt)
            }

            in = survey.Select{Message: "Select build:", Options: listBuilds}
            err = survey.AskOne(&in, &selectedIndex)
            if err != nil {
                if err == terminal.InterruptErr {
                    fmt.Println("Interrupted by user, exiting...")
                    return nil
                }
                log.Fatal(err.Error())
                return err
            }

            selectedBuild := builds[selectedIndex]

            var group *string
            var stream *string
            
            // sometimes when getting log config so quickly it returns empty group or stream
            for _, d := range([]time.Duration{350, 650, 1000}) {
                group = selectedBuild.Logs.GroupName
                stream = selectedBuild.Logs.StreamName

                if group != nil && stream != nil { break }

                time.Sleep(time.Millisecond * d)
                build, err := naws.GetBuildInfo(c, cbClient, buildsListNames[selectedIndex])
                if err != nil {
                    log.Fatal(err.Error())
                    return nil
                }
                selectedBuild = build
                selectedBuild = build
            }

            var token *string
            for  {
                res, err := naws.GetEvents(c, &cfg, group, stream, token)
                if err != nil {
                    log.Fatal(err.Error())
                    return err
                }

                for _, event := range(res.Events) {
                    fmt.Print(*event.Message)
                }

                // check current build state (keeping selectedBuild value)
                build, err := naws.GetBuildInfo(c, cbClient, *selectedBuild.Id)
                if err != nil{
                    log.Fatal(err.Error())
                    return err
                }

                if (token != nil && *token == *res.NextForwardToken && build.BuildComplete) {
                    if !selectedBuild.BuildComplete {
                        beeep.Alert("CodeWatch", "Build completed", "")
                    }
                    break;
                }

                token = res.NextForwardToken
                time.Sleep(2*time.Second)
            }
            
            return nil
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}
