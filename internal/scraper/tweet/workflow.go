package tweet

import (
	"time"

	"github.com/thallesp/go-twitter-scraper/internal/database"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ScrapeTweetsByUsersWorkflow(ctx workflow.Context) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	db := database.GetClientOrPanic()

	var usersIds []string

	result := db.Model(&database.TweetModel{}).Distinct("user_id").Limit(1000).Pluck("user_id", &usersIds)

	if result.Error != nil {
		return result.Error
	}

	if len(usersIds) == 0 {
		workflow.ExecuteActivity(ctx, FindInitialUsers, nil).Get(ctx, nil)
	}

	result = db.Model(&database.TweetModel{}).Distinct("user_id").Limit(1000).Pluck("user_id", &usersIds)

	if result.Error != nil {
		return result.Error
	}

	futureActivities := make([]workflow.Future, len(usersIds))

	for i, userId := range usersIds {
		futureActivities[i] = workflow.ExecuteActivity(ctx, GetTweetsByUserActivity, userId)
	}

	for _, future := range futureActivities {
		future.Get(ctx, nil)
	}

	return nil
}
