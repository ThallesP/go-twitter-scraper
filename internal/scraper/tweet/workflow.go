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

	result := db.Model(&database.TweetModel{}).Distinct("user_id").Limit(1000).Pluck("user_id", &usersIds)

	if result.Error != nil {
		return result.Error
	}

	for _, userId := range usersIds {
		workflow.ExecuteActivity(ctx, GetTweetsByUserActivity, userId)
	}

	return nil
}
