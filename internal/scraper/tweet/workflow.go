package tweet

import (
	"time"

	"github.com/thallesp/go-twitter-scraper/internal/database"
	"go.temporal.io/sdk/workflow"
)

func ScrapeTweetsByUsersWorkflow(ctx workflow.Context) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	db := database.GetClientOrPanic()

	var usersIds []string

	db.Model(&database.TweetModel{}).Distinct("user_id").Where("random() < 0.01").Limit(1000).Pluck("user_id", &usersIds)

	for _, userId := range usersIds {
		workflow.ExecuteActivity(ctx, GetTweetsByUserActivity, userId).Get(ctx, nil)
	}

	return nil
}
