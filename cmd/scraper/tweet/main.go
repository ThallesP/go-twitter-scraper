package main

import (
	"github.com/thallesp/go-twitter-scraper/internal/scraper/tweet"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	tweet.StartUserTweetsScraper()
	tweet.StartUserTweetsWorker()
}
