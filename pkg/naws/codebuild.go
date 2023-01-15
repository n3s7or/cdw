package naws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/urfave/cli/v2"
)

func getClient(c *cli.Context, cfg *aws.Config) *codebuild.Client {
	return codebuild.NewFromConfig(*cfg)
}

func ListBuilds(c *cli.Context, cfg *aws.Config, project *string) ([]string, error){
	client := getClient(c, cfg)	
	return listBuilds(c, client, project)
}

func GetBuildsInfo(c *cli.Context, cfg *aws.Config, ids []string) ([]cbtypes.Build, error) {
	client := getClient(c, cfg)

	res, err := client.BatchGetBuilds(c.Context, &codebuild.BatchGetBuildsInput{Ids: ids})
	if err != nil {
		return []cbtypes.Build{}, err
	}

	return res.Builds, nil
}

func GetBuildInfo(c *cli.Context, cfg *aws.Config, id string) (cbtypes.Build, error) {
	client := getClient(c, cfg)

	res, err := client.BatchGetBuilds(c.Context, &codebuild.BatchGetBuildsInput{Ids: []string{id}})
	if err != nil {
		return cbtypes.Build{}, err
	}

	return res.Builds[0], nil
}


func listBuilds(c *cli.Context, client *codebuild.Client, project *string) ([]string, error){
	res, err := client.ListBuildsForProject(c.Context, &codebuild.ListBuildsForProjectInput{ProjectName: project})
	if err != nil {
		return []string{}, err
	}
	
	return res.Ids, nil
}

func ListProjects(c *cli.Context, cfg *aws.Config, nextToken *string) ([]string, error) {
	client := getClient(c, cfg)
	
	input := codebuild.ListProjectsInput{
		SortBy: "NAME",
		SortOrder: "ASCENDING",
	}

	if nextToken != nil {
		input.NextToken = nextToken
	}

	projects, err := client.ListProjects(c.Context, &input)
	if err != nil {
		return []string{}, err
	}

	return projects.Projects, err
}
