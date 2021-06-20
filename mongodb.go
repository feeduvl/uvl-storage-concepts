package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	database          = "concepts_data"
	collectionDataset = "dataset"
	collectionResult  = "result"

	fieldDatasetName       = "name"
	fieldDatasetUploadedAt = "uploaded_at"
	fieldResultStartedAt   = "started_at"
	fieldResultMethodName  = "method"
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
	datasetIndex := mgo.Index{
		Key:        []string{fieldDatasetName},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	datasetCollection := mongoClient.DB(database).C(collectionDataset)
	err := datasetCollection.EnsureIndex(datasetIndex)
	if err != nil {
		panic(err)
	}
	// Index
	datasetSecondIndex := mgo.Index{
		Key:        []string{fieldDatasetName, fieldDatasetUploadedAt},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	err = datasetCollection.EnsureIndex(datasetSecondIndex)
	if err != nil {
		panic(err)
	}
	// Index
	resultIndex := mgo.Index{
		Key:        []string{fieldResultMethodName, fieldResultStartedAt},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	resultCollection := mongoClient.DB(database).C(collectionResult)
	err = resultCollection.EnsureIndex(resultIndex)
	if err != nil {
		panic(err)
	}
}

// MongoInsertDataset returns ok if the dataset was inserted or already existed
func MongoInsertDataset(mongoClient *mgo.Session, dataset Dataset) error {
	query := bson.M{fieldDatasetName: dataset.Name}
	update := bson.M{"$set": dataset}
	_, err := mongoClient.DB(database).C(collectionDataset).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	}

	return nil
}

// MongoInsertResult returns ok if the result was inserted or already existed
func MongoInsertResult(mongoClient *mgo.Session, result Result) error {
	query := bson.M{fieldResultMethodName: result.Method, fieldResultStartedAt: result.StartedAt}
	update := bson.M{"$set": result}
	_, err := mongoClient.DB(database).C(collectionResult).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	}

	return nil
}

// MongoDeleteDataset return ok if db entry could be deleted
func MongoDeleteDataset(mongoClient *mgo.Session, dataset string) bool {
	_, err := mongoClient.
		DB(database).
		C(collectionDataset).
		RemoveAll(bson.M{fieldDatasetName: dataset})

	return err == nil
}

// MongoDeleteResult return ok if db entry could be deleted
func MongoDeleteResult(mongoClient *mgo.Session, result time.Time) bool {
	_, err := mongoClient.
		DB(database).
		C(collectionResult).
		RemoveAll(bson.M{fieldResultStartedAt: result})

	return err == nil
}

// MongoGetDataset returns a dataset
func MongoGetDataset(mongoClient *mgo.Session, datasetName string) Dataset {
	var dataset []Dataset
	err := mongoClient.
		DB(database).
		C(collectionDataset).
		Find(bson.M{fieldDatasetName: datasetName}).
		All(&dataset)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return dataset[0]
}

// MongoGetAllDatasets returns a dataset
func MongoGetAllDatasets(mongoClient *mgo.Session) []string {

	var datasetNames []string

	err := mongoClient.
		DB(database).
		C(collectionDataset).
		Find(nil).
		Distinct(fieldDatasetName, &datasetNames)

	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	fmt.Printf("getAllDatasets result: %s\n", datasetNames)

	return datasetNames
}

// MongoGetAllResults returns all results
func MongoGetAllResults(mongoClient *mgo.Session) []Result {
	var results []Result
	err := mongoClient.
		DB(database).
		C(collectionResult).
		Find(nil).
		All(&results)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return results
}

/*
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

// MongoGetTweetOfClass returns all tweets belonging to one class i.e., issue report of a specific twitter page
func MongoGetTweetOfClassLimited(mongoClient *mgo.Session, tweetedToName string, tweetClass string, limit int) []Tweet {
	var tweets []Tweet
	err := mongoClient.
		DB(database).
		C(collectionTweet).
		Find(bson.M{fieldInReplyToScreenName: tweetedToName, fieldTweetClass: tweetClass}).
		Limit(limit).
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
*/
