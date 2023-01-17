package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
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
		
		var cbNextToken *string
		projects, err := naws.ListProjects(ctx, &cfg, cbNextToken)
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

		name := strings.Trim(ctx.String("name"), " ")
		if name == ""{
			log.Fatal("no filter string provided")
		}

		cfg, err := config.LoadDefaultConfig(ctx.Context)
            if err != nil {
                log.Fatal(err.Error())
            	return err
            }
		
		projects, err := ListAllProjects(ctx, &cfg)
		if err != nil {
			log.Fatal(err.Error())
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

// ListAllProjects retrieves all projects from codebuild
//  even if there are more than 100 projects
//
// ToDo: check what happens when pagination reaches last page
//
func ListAllProjects(ctx *cli.Context, cfg *aws.Config) ([]string, error) {
	var cbNextToken *string		

	var projects []string

	watchdog := 0

	for {
		res, err := naws.ListProjects(ctx, cfg, cbNextToken)
		if err != nil {
			return []string{}, err
		}
		projects = append(projects, res...)
		
		if cbNextToken == nil { break }	// ToDo: check next token value on the last iteration

		// watchdog to avoid infinite loop until previous ToDo is checked
		// I hope nobody has more than 1k projects
		if watchdog > 9 {
			log.Fatal("KBOOM")
		}
	}

	return projects, nil
}
