package tweet

import (
	"context"

	"github.com/thallesp/go-twitter-scraper/internal/temporal"
	"go.temporal.io/sdk/client"
	temp "go.temporal.io/sdk/temporal"
)

func StartUserTweetsScraper() {
	c := temporal.GetTemporalClientOrFatal()

	options := client.StartWorkflowOptions{
		ID:           "fetch-tweets-by-user",
		TaskQueue:    temporal.ScrapeTweetsByUsersQueue,
		CronSchedule: "*/5 * * * *",
		RetryPolicy: &temp.RetryPolicy{
			NonRetryableErrorTypes: []string{"failed to connect database"},
		},
	}

	c.ExecuteWorkflow(context.Background(), options, ScrapeTweetsByUsersWorkflow)
}
