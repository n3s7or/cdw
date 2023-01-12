package naws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/urfave/cli/v2"
)

func GetClient(c *cli.Context, cfg *aws.Config) *codebuild.Client {
	return codebuild.NewFromConfig(*cfg)
}

func ListBuilds(c *cli.Context, client *codebuild.Client, project *string) ([]string, error){
	builds, err := listBuilds(c, client, project)
	if err != nil {
		return []string{}, err
	}

	return builds, nil
}

func GetBuildsInfo(c *cli.Context, client *codebuild.Client, ids []string) ([]cbtypes.Build, error) {
	res, err := client.BatchGetBuilds(c.Context, &codebuild.BatchGetBuildsInput{Ids: ids})
	if err != nil {
		return []cbtypes.Build{}, err
	}

	return res.Builds, nil
}

func GetBuildInfo(c *cli.Context, client *codebuild.Client, id string) (cbtypes.Build, error) {
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

func ListProjects(c *cli.Context, client *codebuild.Client) ([]string, error) {
	projects, err := client.ListProjects(c.Context, &codebuild.ListProjectsInput{})
	if err != nil {
		return []string{}, err
	}

	return projects.Projects, err
}
