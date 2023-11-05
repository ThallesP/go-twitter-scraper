package tweet

import (
	"math/rand"

	"github.com/thallesp/go-twitter-scraper/internal/database"
)

var cachedAccounts []database.OpenAccountModel

func init() {
	client := database.GetClientOrPanic()

	result := client.Find(&cachedAccounts)

	if result.Error != nil {
		panic(result.Error)
	}
}

func GetRandomAccount() *database.OpenAccountModel {
	randInt := rand.Intn(len(cachedAccounts))

	return &cachedAccounts[randInt]
}
