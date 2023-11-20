package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	_ "embed"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thallesp/go-twitter-scraper/internal/database"
	"gorm.io/gorm"
)

const GuestTokenEndpoint = "https://api.twitter.com/1.1/guest/activate.json"
const FlowTokenEndpoint = "https://api.twitter.com/1.1/onboarding/task.json"
const OpenAccountEndpoint = "https://api.twitter.com/1.1/onboarding/task.json"
const FlowTokenPayload = `{"flow_token":null,"input_flow_data":{"country_code":null,"flow_context":{"start_location":{"location":"splash_screen"}},"requested_variant":null,"target_user_id":0},"subtask_versions":{"generic_urt":3,"standard":1,"open_home_timeline":1,"app_locale_update":1,"enter_date":1,"email_verification":3,"enter_password":5,"enter_text":5,"one_tap":2,"cta":7,"single_sign_on":1,"fetch_persisted_data":1,"enter_username":3,"web_modal":2,"fetch_temporary_password":1,"menu_dialog":1,"sign_up_review":5,"interest_picker":4,"user_recommendations_urt":3,"in_app_notification":1,"sign_up":2,"typeahead_search":1,"user_recommendations_list":4,"cta_inline":1,"contacts_live_sync_permission_prompt":3,"choice_selection":5,"js_instrumentation":1,"alert_dialog_suppress_client_events":1,"privacy_options":1,"topics_selector":1,"wait_spinner":3,"tweet_selection_urt":1,"end_flow":1,"settings_list":7,"open_external_link":1,"phone_verification":5,"security_key":3,"select_banner":2,"upload_media":1,"web":2,"alert_dialog":1,"open_account":2,"action_list":2,"enter_phone":2,"open_link":1,"show_code":1,"update_users":1,"check_logged_in_account":1,"enter_email":2,"select_avatar":4,"location_permission_prompt":2,"notifications_permission_prompt":4}}`
const OpenAccountPayload = `{"flow_token":"||<<REPLACE>>||","subtask_inputs":[{"open_link":{"link":"next_link"},"subtask_id":"NextTaskOpenLink"}],"subtask_versions":{"generic_urt":3,"standard":1,"open_home_timeline":1,"app_locale_update":1,"enter_date":1,"email_verification":3,"enter_password":5,"enter_text":5,"one_tap":2,"cta":7,"single_sign_on":1,"fetch_persisted_data":1,"enter_username":3,"web_modal":2,"fetch_temporary_password":1,"menu_dialog":1,"sign_up_review":5,"interest_picker":4,"user_recommendations_urt":3,"in_app_notification":1,"sign_up":2,"typeahead_search":1,"user_recommendations_list":4,"cta_inline":1,"contacts_live_sync_permission_prompt":3,"choice_selection":5,"js_instrumentation":1,"alert_dialog_suppress_client_events":1,"privacy_options":1,"topics_selector":1,"wait_spinner":3,"tweet_selection_urt":1,"end_flow":1,"settings_list":7,"open_external_link":1,"phone_verification":5,"security_key":3,"select_banner":2,"upload_media":1,"web":2,"alert_dialog":1,"open_account":2,"action_list":2,"enter_phone":2,"open_link":1,"show_code":1,"update_users":1,"check_logged_in_account":1,"enter_email":2,"select_avatar":4,"location_permission_prompt":2,"notifications_permission_prompt":4}}`
const BearerToken = "Bearer AAAAAAAAAAAAAAAAAAAAAFXzAwAAAAAAMHCxpeSDG1gLNLghVe8d74hl6k4%3DRUMF4xAQLsbeBhTSRrCiQpJtxoGWeyHrDb5te2jpGskWDFW82F"

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

func main() {
	client := database.GetClientOrPanic()

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()

			jar, err := cookiejar.New(&cookiejar.Options{})

			if err != nil {
				fmt.Println(err)

				return
			}

			guestToken, err := GenerateGuestToken(jar)

			if err != nil {
				fmt.Println(err)
				return
			}

			flowToken, err := GenerateFlowToken(jar, guestToken)

			if err != nil {
				fmt.Println(err)

				return
			}

			openAccount, err := GenerateOpenAccount(jar, flowToken, guestToken)

			if err != nil {
				fmt.Println(err)
				return
			}

			db.Model(&database.OpenAccountModel{}).Create(&openAccount)
		}(client, wg)
	}

}

func GenerateGuestToken(jar *cookiejar.Jar) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return "", err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Jar:       jar,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GuestTokenEndpoint, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/9.95.0-release.0 (29950000-r-0) ONEPLUS+A3010/9 (OnePlus;ONEPLUS+A3010;OnePlus;OnePlus3;0;;1;2016)")
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

func GenerateFlowToken(jar *cookiejar.Jar, guestToken string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return "", err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Jar:       jar,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", FlowTokenEndpoint, bytes.NewReader([]byte(FlowTokenPayload)))

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/9.95.0-release.0 (29950000-r-0) ONEPLUS+A3010/9 (OnePlus;ONEPLUS+A3010;OnePlus;OnePlus3;0;;1;2016)")
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

func GenerateOpenAccount(jar *cookiejar.Jar, flowToken string, guestToken string) (database.OpenAccountModel, error) {
	proxyUrl, err := url.Parse(os.Getenv("PROXY_ROTATOR_URL"))

	if err != nil {
		return database.OpenAccountModel{}, err
	}

	client := http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Jar:       jar,
	}

	req, err := http.NewRequest("POST", OpenAccountEndpoint, bytes.NewReader([]byte(strings.Replace(OpenAccountPayload, "||<<REPLACE>>||", flowToken, 1))))

	if err != nil {
		return database.OpenAccountModel{}, err
	}

	req.Header.Set("User-Agent", "TwitterAndroid/9.95.0-release.0 (29950000-r-0) ONEPLUS+A3010/9 (OnePlus;ONEPLUS+A3010;OnePlus;OnePlus3;0;;1;2016)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-guest-token", guestToken)
	req.Header.Set("Authorization", BearerToken)

	res, err := client.Do(req)

	if err != nil {
		return database.OpenAccountModel{}, err
	}

	defer res.Body.Close()

	rawOpenAccount := RawOpenAccount{}
	err = json.NewDecoder(res.Body).Decode(&rawOpenAccount)

	if err != nil {
		return database.OpenAccountModel{}, err
	}

	for _, subtask := range rawOpenAccount.Subtasks {
		if subtask.SubtaskId == "OpenAccount" {
			return database.OpenAccountModel{
				AccessToken:       subtask.OpenAccount.OAuthToken,
				AccessTokenSecret: subtask.OpenAccount.OAuthTokenSecret,
			}, nil
		}
	}

	return database.OpenAccountModel{}, errors.New("unable to find OpenAccount subtask")
}
