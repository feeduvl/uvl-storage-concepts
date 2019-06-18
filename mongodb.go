package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	database                    = "twitter_data"
	collectionTweet             = "tweet"
	collectionTwitterProfile    = "twitter_profile"
	collectionObservableTwitter = "observable_twitter"
	collectionTweetLabel        = "tweet_label"
	collectionAccessKeys        = "access_keys"

	fieldInReplyToScreenName = "in_reply_to_screen_name"
	fieldStatusId            = "status_id"
	fieldAccountName         = "account_name"
	fieldLang                = "lang"
	fieldText                = "text"
	fieldUserName            = "user_name"
	fieldProfileName         = "profile_name"
	fieldSentiment           = "sentiment"
	fieldSentimentScore      = "sentiment_score"
	fieldTweetClass          = "tweet_class"
	fieldClassifierCertainty = "classifier_certainty"
	fieldIsAnnotated         = "is_annotated"
	fieldAccessKey           = "access_key"
)

// MongoGetSession returns a session
func MongoGetSession(mongoIP, username, password string) *mgo.Session {
	info := &mgo.DialInfo{
		Addrs:    []string{mongoIP},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatal(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return session
}

// MongoCreateCollectionIndexes creates the indexes
func MongoCreateCollectionIndexes(mongoClient *mgo.Session) {
	// Index
	tweetIndex := mgo.Index{
		Key:        []string{fieldStatusId},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	tweetCollection := mongoClient.DB(database).C(collectionTweet)
	err := tweetCollection.EnsureIndex(tweetIndex)
	if err != nil {
		panic(err)
	}
	// Index
	tweetSecondIndex := mgo.Index{
		Key:        []string{fieldText, fieldUserName},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	err = tweetCollection.EnsureIndex(tweetSecondIndex)
	if err != nil {
		panic(err)
	}

	// Index
	twitterProfileIndex := mgo.Index{
		Key:        []string{fieldProfileName},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	twitterProfileCollection := mongoClient.DB(database).C(collectionTwitterProfile)
	err = twitterProfileCollection.EnsureIndex(twitterProfileIndex)
	if err != nil {
		panic(err)
	}

	// Index
	observableTwitterIndex := mgo.Index{
		Key:        []string{fieldAccountName, fieldLang},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	observableTwitterCollection := mongoClient.DB(database).C(collectionObservableTwitter)
	err = observableTwitterCollection.EnsureIndex(observableTwitterIndex)
	if err != nil {
		panic(err)
	}

	// Index
	tweetLabelIndex := mgo.Index{
		Key:        []string{fieldStatusId},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	tweetLabelCollection := mongoClient.DB(database).C(collectionTweetLabel)
	err = tweetLabelCollection.EnsureIndex(tweetLabelIndex)
	if err != nil {
		panic(err)
	}

	// Index
	accessKeysIndex := mgo.Index{
		Key:        []string{fieldAccessKey},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	accessKeysCollection := mongoClient.DB(database).C(collectionAccessKeys)
	err = accessKeysCollection.EnsureIndex(accessKeysIndex)
	if err != nil {
		panic(err)
	}
}

// MongoInsertTweets returns ok if the tweet was inserted or already existed
func MongoInsertTweets(mongoClient *mgo.Session, tweets []Tweet) bool {
	for _, tweet := range tweets {
		err := mongoClient.DB(database).C(collectionTweet).Insert(tweet)
		if err != nil && !mgo.IsDup(err) {
			fmt.Println(err)
			return false
		}
	}

	return true
}

// MongoUpdateTweetsSentimentAndClass returns ok if the tweet was inserted or already existed
func MongoUpdateTweetsSentimentAndClass(mongoClient *mgo.Session, tweets []Tweet) bool {
	for _, tweet := range tweets {
		query := bson.M{fieldStatusId: tweet.StatusID}
		update := bson.M{"$set": bson.M{
			fieldSentiment:           tweet.Sentiment,
			fieldSentimentScore:      tweet.SentimentScore,
			fieldTweetClass:          tweet.TweetClass,
			fieldClassifierCertainty: tweet.ClassifierCertainty,
		}}
		_, err := mongoClient.DB(database).C(collectionTweet).Upsert(query, update)
		if err != nil && !mgo.IsDup(err) {
			fmt.Println(err)
			return false
		}
	}

	return true
}

// MongoGetTweetOfClass returns all tweets belonging to one class i.e., issue report of a specific twitter page
func MongoGetTweetOfClass(mongoClient *mgo.Session, tweetedToName string, tweetClass string) []Tweet {
	var tweets []Tweet
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Find(bson.M{fieldInReplyToScreenName: tweetedToName, fieldTweetClass: tweetClass}).
		All(&tweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return tweets
}

// MongoGetAllTweetsOfAccountName returns all tweets belonging to one specific twitter page
func MongoGetAllTweetsOfAccountName(mongoClient *mgo.Session, accountName string) []Tweet {
	var tweets []Tweet
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Find(bson.M{fieldInReplyToScreenName: accountName}).
		All(&tweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}
	//, "created_at_full": bson.M{"$exists": true}
	return tweets
}

// MongoGetUnclassifiedAllTweetsOfAccountName returns all tweets belonging to one specific twitter page
func MongoGetUnclassifiedAllTweetsOfAccountName(mongoClient *mgo.Session, accountName, lang string) []Tweet {
	var tweets []Tweet
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Find(bson.M{fieldInReplyToScreenName: accountName, fieldTweetClass: "", fieldLang: lang}).
		All(&tweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}
	//, "created_at_full": bson.M{"$exists": true}
	return tweets
}

// MongoGetAllUnlabeledTweetsOfAccountName returns all tweets of a Twitter account that are not manually labeled yet.
func MongoGetAllUnlabeledTweetsOfAccountName(mongoClient *mgo.Session, accountName string) []Tweet {
	var tweets []Tweet

	var labeledTweets []TweetLabel
	err := mongoClient.
		DB(database).
		C(collectionTweetLabel).
		Find(nil).
		All(&labeledTweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}
	var tweetsToExclude []string
	for _, tweet := range labeledTweets {
		tweetsToExclude = append(tweetsToExclude, tweet.StatusID)
	}

	var query = make(bson.M)
	query[fieldInReplyToScreenName] = accountName
	query[fieldStatusId] = bson.M{"$nin": tweetsToExclude}

	err = mongoClient.
		DB(database).
		C(collectionTweet).
		Find(query).
		All(&tweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return tweets
}

// MongoGetAllTweetsOfAccountForCurrentWeek returns all tweets belonging to one specific twitter page
func MongoGetAllTweetsOfAccountForCurrentWeek(mongoClient *mgo.Session, accountName string, from int, to int) []Tweet {
	var tweets []Tweet
	pipeline := []bson.M{bson.M{
		"$match": bson.M{
			"$and": []bson.M{bson.M{
				fieldInReplyToScreenName: accountName,
				"created_at": bson.M{
					"$gte": from,
					"$lte": to,
				},
			}},
		},
	}}
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Pipe(pipeline).
		All(&tweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return tweets
}

// MongoGetAllTwitterAccounts returns all twitter accounts
func MongoGetAllTwitterAccounts(mongoClient *mgo.Session) TwitterAccount {
	var twitterAccountsRaw []string
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Find(nil).
		Distinct(fieldInReplyToScreenName, &twitterAccountsRaw)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	fmt.Printf("MongoGetAllTwitterAccounts: %v\n", twitterAccountsRaw)

	return TwitterAccount{Names: twitterAccountsRaw}
}

// MongoInsertObservableTwitter returns ok if the package name was inserted or already existed
func MongoInsertObservableTwitter(mongoClient *mgo.Session, observable ObservableTwitter) bool {
	query := bson.M{fieldAccountName: observable.AccountName}
	update := bson.M{"$set": observable}
	_, err := mongoClient.DB(database).C(collectionObservableTwitter).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoGetAllObservableTwitter returns all observable apps
func MongoGetAllObservableTwitter(mongoClient *mgo.Session) []ObservableTwitter {
	var observables []ObservableTwitter
	err := mongoClient.
		DB(database).
		C(collectionObservableTwitter).
		Find(nil).
		All(&observables)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return observables
}

// MongoDeleteObservableTwitter returns ok if db entry could be deleted
func MongoDeleteObservableTwitter(mongoClient *mgo.Session, observable ObservableTwitter) bool {
	_, err := mongoClient.
		DB(database).
		C(collectionObservableTwitter).
		RemoveAll(bson.M{fieldAccountName: observable.AccountName})

	return err == nil
}

// MongoInsertTweetLabel returns ok if the label was inserted or already existed
func MongoInsertTweetLabel(mongoClient *mgo.Session, tweetLabel TweetLabel) bool {
	err := mongoClient.DB(database).C(collectionTweetLabel).Insert(tweetLabel)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}
	return true
}

// MongoUpdateTweetClassAndAnnotation is called when a human provides an annotation for a tweet
func MongoUpdateTweetClassAndAnnotation(mongoClient *mgo.Session, tweetLabel TweetLabel) bool {
	query := bson.M{fieldStatusId: tweetLabel.StatusID}
	update := bson.M{"$set": bson.M{fieldTweetClass: tweetLabel.Label, fieldIsAnnotated: true}}
	_, err := mongoClient.DB(database).C(collectionTweet).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoResetTweetLabels resets the tweet collection
func MongoGetAllLabeledTweets(mongoClient *mgo.Session) []TweetLabel {
	var labeledTweets []TweetLabel
	err := mongoClient.
		DB(database).
		C(collectionTweetLabel).
		Find(nil).
		All(&labeledTweets)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return labeledTweets
}

// MongoUpdateTweetTopics returns ok whether the topics were be updated
func MongoUpdateTweetTopics(mongoClient *mgo.Session, tweet Tweet) bool {
	query := bson.M{"status_id": tweet.StatusID}
	update := bson.M{"$set": bson.M{"topics": tweet.Topics}}
	_, err := mongoClient.DB(database).C(collectionTweet).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoGetAccessKeyExists returns true if the key is in the database
func MongoGetAccessKeyExists(mongoClient *mgo.Session, accessKey AccessKey) bool {
	count, err := mongoClient.
		DB(database).
		C(collectionAccessKeys).
		Find(bson.M{fieldAccessKey: accessKey.Key}).
		Count()
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return count > 0
}

// MongoGetAccessKeyConfiguration returns true if the key is in the database
func MongoGetAccessKeyConfiguration(mongoClient *mgo.Session, accessKey AccessKey) AccessKeyConfiguration {
	var accessKeyDB AccessKey
	err := mongoClient.
		DB(database).
		C(collectionAccessKeys).
		Find(bson.M{fieldAccessKey: accessKey.Key}).
		One(&accessKeyDB)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return accessKeyDB.Configuration
}

// MongoUpdateAccessKeyConfiguration
func MongoUpdateAccessKeyConfiguration(mongoClient *mgo.Session, accessKey AccessKey) {
	query := bson.M{fieldAccessKey: accessKey.Key}
	update := bson.M{"$set": bson.M{
		"configuration.twitter_accounts":           accessKey.Configuration.TwitterAccounts,
		"configuration.google_play_store_accounts": accessKey.Configuration.GooglePlayStoreAccounts,
		"configuration.topics":                     accessKey.Configuration.Topics,
	}}
	_, err := mongoClient.DB(database).C(collectionAccessKeys).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
	}
}
