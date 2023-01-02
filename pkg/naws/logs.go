package naws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/urfave/cli/v2"
)


func GetEvents(c *cli.Context, cfg *aws.Config, group *string, stream *string, nextToken *string) (*cloudwatchlogs.GetLogEventsOutput, error){
	client := cloudwatchlogs.NewFromConfig(*cfg)

	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName: group, 
		LogStreamName: stream,
		StartFromHead: aws.Bool(true),		// to make next token work 
	}

	if nextToken != nil {
		input.NextToken = nextToken
	}
	
	res, err := client.GetLogEvents(c.Context, input)
	if err != nil {
		return &cloudwatchlogs.GetLogEventsOutput{}, err
	}

	return res, nil
}
