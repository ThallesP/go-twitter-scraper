package temporal

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
)

func GetTemporalClientOrFatal() client.Client {
	c, err := client.Dial(client.Options{
		HostPort: os.Getenv("TEMPORAL_HOST_PORT"),
	})

	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}

	return c
}

const ScrapeTweetsByUsersQueue = "SCRAPE_TWEETS_BY_USERS_TASK_QUEUE"
