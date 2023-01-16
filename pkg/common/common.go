package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/gen2brain/beeep"
	"github.com/n3s7or/cdw/pkg/naws"
	"github.com/urfave/cli/v2"
)

func MonitorBuild(c *cli.Context, cfg *aws.Config, build *cbtypes.Build) error{
	fmt.Printf("Showing logs for build: `%s`\n\n", *build.Id)
	
	var group *string
	var stream *string

	// sometimes when getting log config so quickly it returns empty group or stream
	for _, d := range([]time.Duration{350, 650, 1000}) {
		group = build.Logs.GroupName
		stream = build.Logs.StreamName

		if group != nil && stream != nil { break }

		time.Sleep(time.Millisecond * d)
		buildTemp, err := naws.GetBuildInfo(c, cfg, *build.Id)
		if err != nil {
			return err
		}
		*build = buildTemp
		*build = buildTemp
	}

	var token *string
	for  {
		res, err := naws.GetEvents(c, cfg, group, stream, token)
		if err != nil {
			return err
		}

		for _, event := range(res.Events) {
			fmt.Print(*event.Message)
		}

		buildTemp, err := naws.GetBuildInfo(c, cfg, *build.Id)
		if err != nil{
			return err
		}

		if (token != nil && *token == *res.NextForwardToken && buildTemp.BuildComplete) {
			if !build.BuildComplete {
				beeep.Alert("CodeWatch", "Build completed", "")
			}

			return nil
		}

		token = res.NextForwardToken
		time.Sleep(2*time.Second)
	}
}

func MonitorBuildState(c *cli.Context, cfg *aws.Config, build *cbtypes.Build) error {
	fmt.Printf("Checking build with id: %s\n\n", *build.Id)

	var printFrom int

	for {
		build, err := naws.GetBuildInfo(c, cfg, *build.Id)
		if err != nil{
			return err
		}

		phasesCount := len(build.Phases)
		for _, phase := range(build.Phases[printFrom: phasesCount]) {
			fmt.Printf("%s\n", phase.PhaseType)
		}
		printFrom = phasesCount
		
		if (build.BuildComplete) {
			if !build.BuildComplete {
				beeep.Alert("CodeWatch", "Build completed", "")
			}

			fmt.Printf("\nBuild completed: %s\n", build.BuildStatus)
			return nil
		}

		time.Sleep(1500 * time.Millisecond)
	}
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
