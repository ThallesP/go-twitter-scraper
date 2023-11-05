package tweet

import (
	"log"

	"github.com/thallesp/go-twitter-scraper/internal/temporal"
	"go.temporal.io/sdk/worker"
)

func StartUserTweetsWorker() {
	c := temporal.GetTemporalClientOrFatal()

	defer c.Close()

	w := worker.New(c, temporal.ScrapeTweetsByUsersQueue, worker.Options{})

	w.RegisterWorkflow(ScrapeTweetsByUsersWorkflow)
	w.RegisterActivity(GetTweetsByUserActivity)

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
