package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

var ProjectsCommand = cli.Command{
	Name: "projects",
	Usage: "List up to 100 projects, if your account has more projects use filter",
	Subcommands: []*cli.Command{&ProjectsFilterCommand},
	Action: func(ctx *cli.Context) error {

		cfg, err := config.LoadDefaultConfig(ctx.Context)
            if err != nil {
                log.Fatal(err.Error())
            	return err
            }
		
		cbClient := naws.GetClient(ctx, &cfg)
		
		var cbNextToken *string // ToDo: get other projects if there is a next token
		projects, err := naws.ListProjects(ctx, cbClient, cbNextToken)
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		fmt.Printf("%s\n", strings.Join(projects, "\n"))

		return nil
	},
}

var ProjectsFilterCommand = cli.Command{
	Name: "filter",
	Usage: "Filters projects by name",
	Flags: []cli.Flag{&cli.StringFlag{Name: "name", Required: true, Usage: "string to search"}},
	Action: func(ctx *cli.Context) error {

		name := ctx.String("name")
		if name == ""{
			log.Fatal("no filter string provided")
		}

		cfg, err := config.LoadDefaultConfig(ctx.Context)
            if err != nil {
                log.Fatal(err.Error())
            	return err
            }
		
		cbClient := naws.GetClient(ctx, &cfg)

		var cbNextToken *string
		var projects []string

		wd := 0

		for {
			res, err := naws.ListProjects(ctx, cbClient, cbNextToken)
			if err != nil {
				log.Fatal(err.Error())
				return err
			}
			projects = append(projects, res...)
			
			// ToDo: check next token value on the last iteration
			if cbNextToken == nil {
				break
			}

			// watchdog to avoid infinite loop until previous todo is fixed
			// I hope nobody has more than 1k projects
			if wd > 9 {
				log.Fatal("KBOOM")
				break
			}
		}

		// todo: create a generic filter function for cases like this
		var filteredProjects []string
		for _, item := range(projects){
			if strings.Contains(item, name) {
				filteredProjects = append(filteredProjects, item)
			}
		}
		
		fmt.Printf("%s\n", strings.Join(filteredProjects, "\n"))		

		return nil
	},
}
