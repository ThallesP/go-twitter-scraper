package database

type OpenAccountModel struct {
	ID                int    `gorm:"primaryKey" json:"id"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

type TweetModel struct {
	ID            string `gorm:"primaryKey" json:"id"`
	Content       string `json:"content"`
	UserID        string `json:"user_id"`
	RetweetCount  int    `json:"retweet_count"`
	FavoriteCount int    `json:"favorite_count"`
	Lang          string `json:"lang"`
	CreatedAt     string `json:"created_at"`
}
