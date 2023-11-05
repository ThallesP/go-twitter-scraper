package main

import (
	"github.com/thallesp/go-twitter-scraper/internal/scraper/tweet"
	"github.com/thallesp/go-twitter-scraper/internal/utils"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	handler := utils.SetupLogger()
	defer handler.Close()

	tweet.StartUserTweetsScraper()
	tweet.StartUserTweetsWorker()
}
