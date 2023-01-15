package common

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/gen2brain/beeep"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

func MonitorBuild(c *cli.Context, cfg *aws.Config, selectedBuild *cbtypes.Build) error {
	fmt.Printf("Showing logs for build: `%s`\n\n", *selectedBuild.Id)
	
	var group *string
	var stream *string

	// sometimes when getting log config so quickly it returns empty group or stream
	for _, d := range([]time.Duration{350, 650, 1000}) {
		group = selectedBuild.Logs.GroupName
		stream = selectedBuild.Logs.StreamName

		if group != nil && stream != nil { break }

		time.Sleep(time.Millisecond * d)
		build, err := naws.GetBuildInfo(c, cfg, *selectedBuild.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
		*selectedBuild = build
		*selectedBuild = build
	}

	var token *string
	for  {
		res, err := naws.GetEvents(c, cfg, group, stream, token)
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, event := range(res.Events) {
			fmt.Print(*event.Message)
		}

		// check current build state (keeping selectedBuild value)
		build, err := naws.GetBuildInfo(c, cfg, *selectedBuild.Id)
		if err != nil{
			log.Fatal(err.Error())
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
}

func GetBuildsFromBuildsName(c *cli.Context, cfg aws.Config, buildsName []string) ([]cbtypes.Build, error) {
	if len(buildsName) > 9 {
		return naws.GetBuildsInfo(c, &cfg, buildsName[0:10])
	}else {
		return naws.GetBuildsInfo(c, &cfg, buildsName)
	}
}

func GenerateOptionsListForsurvey(builds []cbtypes.Build) ([]string) {
	var listBuilds []string;
	for _, item := range(builds){
		id := string(strings.Split((*item.Id), ":")[1])
		initiator := string(*item.Initiator)
		opt := fmt.Sprintf("%-*s\t%s\t(%s)", 11, fmt.Sprintf("[%s]", string(item.BuildStatus)), initiator, id)
		listBuilds = append(listBuilds, opt)
	}
	return listBuilds
}
