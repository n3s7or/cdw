package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/config"
    cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Action: func(c *cli.Context) error {

            cfg, err := config.LoadDefaultConfig(c.Context)
            if err != nil {
            	return err
            }

            cbClient := naws.GetClient(c, &cfg)

            projects, err := naws.ListProjects(c, cbClient)
            if err != nil {
                println(err.Error())
                return err
            }
            
            in := survey.Select{Message: "Select project:", Options: projects}
            var selectedIndex int
            survey.AskOne(&in, &selectedIndex)

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
                println(err.Error())
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
            survey.AskOne(&in, &selectedIndex)

            group := builds[selectedIndex].Logs.GroupName
            stream := builds[selectedIndex].Logs.StreamName

            var token *string
            for  {
                res, err := naws.GetEvents(c, &cfg, group, stream, token)
                if err != nil {
                    return err
                }

                for _, event := range(res.Events) {
                    fmt.Print(*event.Message)
                }

                // check current build state
                build, err := naws.GetBuildsInfo(c, cbClient, buildsListNames[selectedIndex:selectedIndex+1])
                if err != nil{
                    return err
                }

                if (token != nil && *token == *res.NextForwardToken && build[0].BuildComplete) {
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
