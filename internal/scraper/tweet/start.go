package tweet

import (
	"context"

	"github.com/thallesp/go-twitter-scraper/internal/temporal"
	"go.temporal.io/sdk/client"
)

func StartUserTweetsScraper() {
	c := temporal.GetTemporalClientOrFatal()

	options := client.StartWorkflowOptions{
		ID:           "fetch-tweets-by-user",
		TaskQueue:    temporal.ScrapeTweetsByUsersQueue,
		CronSchedule: "0 * * * *", // every 1 hour
	}

	c.ExecuteWorkflow(context.Background(), options, ScrapeTweetsByUsersWorkflow)
}
