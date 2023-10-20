package twitter_scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"golang.org/x/exp/maps"
)

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

type RawGuestToken struct {
	GuestToken string `json:"guest_token"`
}

type RawFlowToken struct {
	FlowToken string `json:"flow_token"`
}

type RawOpenAccount struct {
	Subtasks []struct {
		SubtaskId   string `json:"subtask_id"`
		OpenAccount struct {
			OAuthToken       string `json:"oauth_token"`
			OAuthTokenSecret string `json:"oauth_token_secret"`
		} `json:"open_account"`
	} `json:"subtasks"`
}

func GetTweetsByUser(ctx context.Context, userId string) ([]Tweet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Printf("Fetching tweets for user %s", userId)
	token := oauth1.NewToken("1710518736753442816-HGVdJOQ6WFAL31mLH6Ui9ryKErp6ZW", "Hsb8K6iHAuNE3FOh9rfolMxTN1MJd7QreqTPqFVRsiemj")

	httpClient := OAuthConfig.Client(ctx, token)

	req, err := http.NewRequest("GET", LegacyUsersTweetsEndpoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")

	q := req.URL.Query()
	q.Add("id", userId)
	q.Add("count", "100")
	q.Add("tweet_mode", "extended")
	req.URL.RawQuery = q.Encode()

	res, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var rawTweets RawResponse

	err = json.NewDecoder(res.Body).Decode(&rawTweets)

	if err != nil {
		return nil, err
	}

	v := maps.Values(rawTweets.TwitterObjects.Tweets)

	log.Println(v)

	return nil, err
}

func GenerateGuestToken() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return "", err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GuestTokenEndpoint, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")
	req.Header.Set("Authorization", BearerToken)

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", errors.New("unable to generate guest token received status code " + res.Status + " from Twitter")
	}

	defer res.Body.Close()

	rawGuestToken := RawGuestToken{}

	err = json.NewDecoder(res.Body).Decode(&rawGuestToken)

	if err != nil {
		return "", err
	}

	return rawGuestToken.GuestToken, nil
}

func GenerateFlowToken(guestToken string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return "", err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", FlowTokenEndpoint, bytes.NewReader([]byte(FlowTokenPayload)))

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-guest-token", guestToken)
	req.Header.Set("Authorization", BearerToken)

	q := req.URL.Query()
	q.Add("flow_name", "welcome")
	q.Add("api_version", "1")
	q.Add("known_device_token", "")
	q.Add("sim_country_code", "us")

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	rawFlowToken := RawFlowToken{}

	err = json.NewDecoder(res.Body).Decode(&rawFlowToken)

	if err != nil {
		return "", err
	}

	return rawFlowToken.FlowToken, nil
}

func GenerateOpenAccount(flowToken string, guestToken string) (OpenAccount, error) {
	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return OpenAccount{}, err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}

	req, err := http.NewRequest("POST", OpenAccountEndpoint, bytes.NewReader([]byte(strings.Replace(OpenAccountPayload, "||<<REPLACE>>||", flowToken, 1))))

	if err != nil {
		return OpenAccount{}, err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/99")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-guest-token", guestToken)
	req.Header.Set("Authorization", BearerToken)

	res, err := client.Do(req)

	if err != nil {
		return OpenAccount{}, err
	}

	defer res.Body.Close()

	rawOpenAccount := RawOpenAccount{}
	err = json.NewDecoder(res.Body).Decode(&rawOpenAccount)

	if err != nil {
		return OpenAccount{}, err
	}

	for _, subtask := range rawOpenAccount.Subtasks {
		if subtask.SubtaskId == "OpenAccount" {
			return OpenAccount{
				AccessToken:       subtask.OpenAccount.OAuthToken,
				AccessTokenSecret: subtask.OpenAccount.OAuthTokenSecret,
			}, nil
		}
	}

	return OpenAccount{}, errors.New("unable to find OpenAccount subtask")
}
