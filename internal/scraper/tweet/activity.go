package tweet

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/thallesp/go-twitter-scraper/internal/database"
	"golang.org/x/exp/maps"
	"gorm.io/gorm/clause"
)

var OAuthConfig = oauth1.Config{
	ConsumerKey:    "3nVuSoBZnx6U4vzUxf5w",
	ConsumerSecret: "Bcs59EFbbsdF6Sl9Ng71smgStWEGwXXKSjYvPVt7qys",
}

const LegacyUsersTweetsEndpoint = "https://api.twitter.com/1.1/timeline/user.json"

type RawTweet struct {
	CreatedAt string `json:"created_at"`
	ID        int64  `json:"id"`
	IDStr     string `json:"id_str"`
	FullText  string `json:"full_text"`
	User      struct {
		ID                 int    `json:"id"`
		IDStr              string `json:"id_str"`
		HasNoScreenName    bool   `json:"has_no_screen_name"`
		RequireSomeConsent bool   `json:"require_some_consent"`
	} `json:"user"`
	RetweetCount  int    `json:"retweet_count"`
	FavoriteCount int    `json:"favorite_count"`
	Lang          string `json:"lang"`
}

type RawResponse struct {
	TwitterObjects struct {
		Tweets map[string]RawTweet `json:"tweets"`
	} `json:"twitter_objects"`
}

type RawResponseSearch struct {
	Metadata struct {
		Cursor               string `json:"cursor"`
		RefreshIntervalInSec int    `json:"refresh_interval_in_sec"`
	} `json:"metadata"`
	Modules []struct {
		Status struct {
			Data struct {
				CreatedAt     string `json:"created_at"`
				IDStr         string `json:"id_str"`
				FullText      string `json:"full_text"`
				RetweetCount  int    `json:"retweet_count"`
				FavoriteCount int    `json:"favorite_count"`
				User          struct {
					IDStr string `json:"id_str"`
				} `json:"user"`
				Lang string `json:"lang"`
			} `json:"data"`
		} `json:"status"`
	} `json:"modules"`
}

func GetTweetsByUserActivity(ctx context.Context, userId string) error {
	slog.Info("Fetching tweets from user", slog.String("user_id", userId))
	account := GetRandomAccount()
	token := oauth1.NewToken(account.AccessToken, account.AccessTokenSecret)

	httpClient := OAuthConfig.Client(ctx, token)

	req, err := http.NewRequest("GET", LegacyUsersTweetsEndpoint, nil)

	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")

	q := req.URL.Query()
	q.Add("id", userId)
	q.Add("count", "100")
	q.Add("tweet_mode", "extended")
	req.URL.RawQuery = q.Encode()

	res, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	var rawResponse RawResponse

	json.NewDecoder(res.Body).Decode(&rawResponse)

	tweets := maps.Values(rawResponse.TwitterObjects.Tweets)

	if len(tweets) == 0 {
		return nil
	}

	client := database.GetClientOrPanic()

	tweetsInsert := make([]database.TweetModel, len(tweets))

	for i, tweet := range tweets {
		tweetsInsert[i] = database.TweetModel{
			ID:            tweet.IDStr,
			Content:       tweet.FullText,
			UserID:        tweet.User.IDStr,
			CreatedAt:     tweet.CreatedAt,
			RetweetCount:  tweet.RetweetCount,
			FavoriteCount: tweet.FavoriteCount,
			Lang:          tweet.Lang,
		}
	}

	slog.Info("Inserting tweets", slog.Int("count", len(tweetsInsert)))

	result := client.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&tweetsInsert)

	if result.Error != nil {
		return result.Error
	}

	slog.Info("Inserted tweets", slog.Int("count", len(tweetsInsert)))

	return nil
}

func FindInitialUsers(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	account := GetRandomAccount()

	token := oauth1.NewToken(account.AccessToken, account.AccessTokenSecret)

	httpClient := OAuthConfig.Client(ctx, token)

	req, err := http.NewRequest("GET", "https://api.twitter.com/1.1/search/universal.json", nil)

	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")

	q := req.URL.Query()
	q.Add("q", "a")
	q.Add("result_type", "recent")
	q.Add("count", "20")
	q.Add("tweet_mode", "extended")
	req.URL.RawQuery = q.Encode()

	res, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	var rawResponseSearch RawResponseSearch
	err = json.NewDecoder(res.Body).Decode(&rawResponseSearch)

	if err != nil {
		return err
	}

	client := database.GetClientOrPanic()

	for _, tweet := range rawResponseSearch.Modules {
		client.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&database.TweetModel{
			ID:            tweet.Status.Data.IDStr,
			Content:       tweet.Status.Data.FullText,
			UserID:        tweet.Status.Data.User.IDStr,
			RetweetCount:  tweet.Status.Data.RetweetCount,
			Lang:          tweet.Status.Data.Lang,
			FavoriteCount: tweet.Status.Data.FavoriteCount,
			CreatedAt:     tweet.Status.Data.CreatedAt,
		})
	}

	return nil
}
