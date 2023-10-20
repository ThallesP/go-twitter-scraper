package twitter_scraper

import (
	"log"
	"time"

	"go.temporal.io/sdk/workflow"
)

func FetchTweetsByUserWorkflow(ctx workflow.Context, userId string) error {

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	tweetsErr := workflow.ExecuteActivity(ctx, GetTweetsByUser, userId).Get(ctx, nil)

	if tweetsErr != nil {
		return tweetsErr
	}

	return nil
}

func GenerateOpenAccountWorkflow(ctx workflow.Context) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var guestToken string
	var flowToken string
	var openAccount OpenAccount
	guestTokenErr := workflow.ExecuteActivity(ctx, GenerateGuestToken).Get(ctx, &guestToken)

	if guestTokenErr != nil {
		return guestTokenErr
	}

	flowTokenErr := workflow.ExecuteActivity(ctx, GenerateFlowToken, guestToken).Get(ctx, &flowToken)

	if flowTokenErr != nil {
		return flowTokenErr
	}

	openAccountErr := workflow.ExecuteActivity(ctx, GenerateOpenAccount, flowToken, guestToken).Get(ctx, &openAccount)

	if openAccountErr != nil {
		return flowTokenErr
	}

	log.Println(guestToken)
	log.Println(flowToken)
	log.Println(openAccount)

	return nil
}
