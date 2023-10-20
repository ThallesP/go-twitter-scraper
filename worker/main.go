package main

import (
	"log"

	"github.com/thallesp/twitter_scraper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: "192.168.0.13:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client.", err)
	}
	defer c.Close()

	//w := worker.New(c, twitter_scraper.FetchTweetsByUserQueueName, worker.Options{})

	// This worker hosts both Workflow and Activity functions.
	//w.RegisterWorkflow(twitter_scraper.FetchTweetsByUserWorkflow)
	//w.RegisterActivity(twitter_scraper.GetTweetsByUser)
	wopen := worker.New(c, twitter_scraper.GenerateOpenAccountQueueName, worker.Options{})

	wopen.RegisterWorkflow(twitter_scraper.GenerateOpenAccountWorkflow)
	wopen.RegisterActivity(twitter_scraper.GenerateGuestToken)
	wopen.RegisterActivity(twitter_scraper.GenerateFlowToken)
	wopen.RegisterActivity(twitter_scraper.GenerateOpenAccount)

	// Start listening to the Task Queue.
	err = wopen.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
