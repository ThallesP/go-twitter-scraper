package main

import (
	"context"
	"log"

	"github.com/thallesp/twitter_scraper"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: "192.168.0.13:7233",
	})

	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}

	defer c.Close()

	ExecuteGenerateOpenAccount(c)
	//ExecuteFetchTweetsByUserWorkflow(c)
}

func ExecuteGenerateOpenAccount(c client.Client) {
	options := client.StartWorkflowOptions{
		ID:        "generate-open-account",
		TaskQueue: twitter_scraper.GenerateOpenAccountQueueName,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, twitter_scraper.GenerateOpenAccountWorkflow)
	if err != nil {
		log.Fatalln("Unable to start the Workflow:", err)
	}

	err = we.Get(context.Background(), nil)

	if err != nil {
		log.Fatalln("Unable to get Workflow result:", err)
	}
}

func ExecuteFetchTweetsByUserWorkflow(c client.Client) error {
	options := client.StartWorkflowOptions{
		ID:        "fetch-tweets-by-user-44196397",
		TaskQueue: twitter_scraper.FetchTweetsByUserQueueName,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, twitter_scraper.FetchTweetsByUserWorkflow, "44196397")
	if err != nil {
		log.Fatalln("Unable to start the Workflow:", err)
	}

	err = we.Get(context.Background(), nil)

	if err != nil {
		log.Fatalln("Unable to get Workflow result:", err)
	}

	return nil
}
