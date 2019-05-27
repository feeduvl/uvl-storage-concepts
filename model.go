package main

import validator "gopkg.in/validator.v2"

// Tweet model
type Tweet struct {
	CreatedAt           int         `validate:"nonzero" json:"created_at" bson:"created_at"`
	CreatedAtFull       string      `json:"created_at_full" bson:"created_at_full"`
	FavoriteCount       int         `json:"favorite_count" bson:"favorite_count"`
	RetweetCount        int         `json:"retweet_count" bson:"retweet_count"`
	Text                string      `validate:"nonzero" json:"text" bson:"text"`
	StatusID            string      `validate:"nonzero" json:"status_id" bson:"status_id"`
	UserName            string      `json:"user_name" bson:"user_name"`
	InReplyToScreenName string      `json:"in_reply_to_screen_name" bson:"in_reply_to_screen_name"`
	Hashtags            []string    `json:"hashtags" bson:"hashtags"`
	Lang                string      `json:"lang" bson:"lang"`
	Sentiment           string      `json:"sentiment" bson:"sentiment"`
	SentimentScore      int         `json:"sentiment_score" bson:"sentiment_score"`
	TweetClass          string      `json:"tweet_class" bson:"tweet_class"`
	ClassifierCertainty int         `json:"classifier_certainty" bson:"classifier_certainty"`
	Annotated           bool        `json:"is_annotated" bson:"is_annotated"`
	Topics              TweetTopics `json:"topics" bson:"topics"`
}

func (tweet *Tweet) validate() error {
	return validator.Validate(tweet)
}

func validateTweets(tweets []Tweet) error {
	for _, tweet := range tweets {
		err := tweet.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// TweetLabel model
type TweetLabel struct {
	StatusID      string `validate:"nonzero" json:"status_id" bson:"status_id"`
	Date          int    `json:"date" bson:"date"` // formt: YYYYmmmdd
	Label         string `validate:"nonzero" json:"label" bson:"label"`
	PreviousLabel string `json:"previous_label" bson:"previous_label"`
}

func (tweetLabel *TweetLabel) validate() error {
	return validator.Validate(tweetLabel)
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

// TwitterAccount model
type TwitterAccount struct {
	Names []string `json:"twitter_account_names" bson:"twitter_account_names"`
}

type TweetTopics struct {
	FirstClass struct {
		Label string  `json:"label" bson:"label"`
		Score float64 `json:"score" bson:"score"`
	} `json:"first_class" bson:"first_class"`
	SecondClass struct {
		Label string  `json:"label" bson:"label"`
		Score float64 `json:"score" bson:"score"`
	} `json:"second_class" bson:"second_class"`
}

// ObservableTwitter model
type ObservableTwitter struct {
	AccountName string `validate:"nonzero" json:"account_name" bson:"account_name"`
	Interval    string `json:"interval" bson:"interval"`
	Lang        string `validate:"nonzero" json:"lang" bson:"lang"`
}

func (observable *ObservableTwitter) validate() error {
	return validator.Validate(observable)
}

type AccessKey struct {
	Key           string                 `validate:"nonzero" json:"access_key" bson:"access_key"`
	Configuration AccessKeyConfiguration `json:"configuration" bson:"configuration"`
}

type AccessKeyConfiguration struct {
	TwitterAccounts         []string `json:"twitter_accounts" bson:"twitter_accounts"`
	GooglePlayStoreAccounts []string `json:"google_play_store_accounts" bson:"google_play_store_accounts"`
	Topics                  []string `json:"topics" bson:"topics"`
}
