package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/dbtest"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gorilla/mux"
	"testing"
)

var router *mux.Router
var mockDBServer dbtest.DBServer
var documents []Document

/*var tweets []Tweet
var invalidTweet Tweet
var existingAccessKey AccessKey
var notExistingAccessKey AccessKey
var invalidArrayPayload []byte
var invalidObjectPayload []byte*/

func TestMain(m *testing.M) {
	fmt.Println("--- Start Tests")
	setup()

	// run the test cases defined in this file
	retCode := m.Run()

	tearDown()

	// call with result of m.Run()
	os.Exit(retCode)
}

func setup() {
	fmt.Println("--- --- setup")
	setupRouter()
	setupDB()
	fillDB()
}

func setupRouter() {
	router = makeRouter()
}

func setupDB() {
	mongoClient = MongoGetSession(os.Getenv("MONGO_IP"), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), databaseTest)
	MongoCreateCollectionIndexes(mongoClient)
}

func fillDB() {

	/*documents = append(documents, Document {
		Id: 	"0",
		Text: "Text 1",
	})
	documents = append(documents, Document {
		Id: 	"1",
		Text: "Text 2",
	})
	documents = append(documents, Document {
		Id: 	"2",
		Text: "Text 3",
	})

	/*
	 * Insert fake datasets
	*/
	/*
		err := mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
			UploadedAt:      time.Now(),
			Name:         	"test_dataset_1",
			Size: 			3,
			Documents:      documents,
		})
		if err != nil {
			panic(err)
		}

		err = mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
			UploadedAt:      time.Now(),
			Name:         	"test_dataset_2",
			Size: 			3,
			Documents:      documents,
		})
		if err != nil {
			panic(err)
		}

		err = mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
			UploadedAt:      time.Now(),
			Name:         	"test_dataset_3",
			Size: 			3,
			Documents:      documents,
		})
		if err != nil {
			panic(err)
		}*/

	/*
		 * Insert fake tweets

		fmt.Println("Insert fake tweets")
		tweets = append(tweets, Tweet{
			CreatedAt:           20180121,
			CreatedAtFull:       "Mon Jan 21 12:28:28 +0000 2019",
			FavoriteCount:       0,
			RetweetCount:        0,
			Text:                "@Tre_It complimenti per Luca, un vostro collaboratore che lavora presso MediaWord di Cinisello Balsamo. Una persona attenta, precisa e sempre disponibile nei confronti dei clienti. #bellapersona",
			StatusID:            "1",
			UserName:            "nytwitt",
			InReplyToScreenName: "Tre_It",
			Hashtags:            []string{"bellapersona"},
			Lang:                "it",
			Sentiment:           "NEUTRAL",
			SentimentScore:      0,
			TweetClass:          "irrelevant",
			ClassifierCertainty: 0,
		})
		tweets = append(tweets, Tweet{
			CreatedAt:           20180121,
			CreatedAtFull:       "Mon Jan 21 12:28:28 +0000 2019",
			FavoriteCount:       0,
			RetweetCount:        0,
			Text:                "@WindItalia complimenti per Luca, un vostro collaboratore che lavora presso MediaWord di Cinisello Balsamo. Una persona attenta, precisa e sempre disponibile nei confronti dei clienti. #bellapersona",
			StatusID:            "2",
			UserName:            "katast",
			InReplyToScreenName: "WindItalia",
			Hashtags:            []string{"bellapersona"},
			Lang:                "it",
			Sentiment:           "NEUTRAL",
			SentimentScore:      0,
			TweetClass:          "problem_report",
			ClassifierCertainty: 0,
		})
		tweets = append(tweets, Tweet{
			CreatedAt:           20180121,
			CreatedAtFull:       "Mon Jan 21 12:28:28 +0000 2019",
			FavoriteCount:       0,
			RetweetCount:        0,
			Text:                "@Tre_It complimenti per Luca, un vostro collaboratore che lavora presso MediaWord di Cinisello Balsamo. Una persona attenta, precisa e sempre disponibile nei confronti dei clienti. #bellapersona",
			StatusID:            "3",
			UserName:            "creat",
			InReplyToScreenName: "Tre_It",
			Hashtags:            []string{"bellapersona"},
			Lang:                "it",
			Sentiment:           "NEUTRAL",
			SentimentScore:      0,
			TweetClass:          "",
			ClassifierCertainty: 0,
		})
		dateOfCurrentWeek, _ := strconv.Atoi(time.Now().AddDate(0, 0, -5).Format("20060102"))
		tweets = append(tweets, Tweet{
			CreatedAt:           dateOfCurrentWeek,
			CreatedAtFull:       "Mon Jan 21 12:28:28 +0000 2019",
			FavoriteCount:       0,
			RetweetCount:        0,
			Text:                "@Tre_It complimenti per Luca, un vostro collaboratore che lavora presso MediaWord di Cinisello Balsamo. Una persona attenta, precisa e sempre disponibile nei confronti dei clienti. #bellapersona",
			StatusID:            "4",
			UserName:            "charl",
			InReplyToScreenName: "Tre_It",
			Hashtags:            []string{"bellapersona"},
			Lang:                "it",
			Sentiment:           "NEUTRAL",
			SentimentScore:      0,
			TweetClass:          "inquiry",
			ClassifierCertainty: 0,
		})

		tweetBulk := mongoClient.DB(database).C(collectionTweet).Bulk()
		for _, tweet := range tweets {
			tweetBulk.Insert(tweet)
		}
		_, err := tweetBulk.Run()
		if err != nil {
			panic(err)
		}

		invalidTweet = Tweet{}
		invalidArrayPayload = []byte(`[{ "wrong_json_format": true }]`)
		invalidObjectPayload = []byte(`{ "wrong_json_format": true }`)

		notExistingAccessKey = AccessKey{
			Key: "notindb",
			Configuration: AccessKeyConfiguration{
				TwitterAccounts:         []string{"Tre_It"},
				GooglePlayStoreAccounts: []string{},
				Topics:                  []string{"Network"},
			},
		}

		existingAccessKey = AccessKey{
			Key: "indb",
			Configuration: AccessKeyConfiguration{
				TwitterAccounts:         []string{"Tre_It"},
				GooglePlayStoreAccounts: []string{"Wind Tre S.p.A."},
				Topics:                  []string{"Contract", "Devices"},
			},
		}

		err = mongoClient.DB(database).C(collectionAccessKeys).Insert(existingAccessKey)
		if err != nil {
			panic(err)
		}

	*/

	/*
		 * Insert fake observables

		err = mongoClient.DB(database).C(collectionObservableTwitter).Insert(ObservableTwitter{
			AccountName: "TestObserverable",
			Interval:    "2h",
			Lang:        "en",
		})
		if err != nil {
			panic(err)
		}
	*/
	/*
		 * Insert fake labels

		err = mongoClient.DB(database).C(collectionTweetLabel).Insert(TweetLabel{
			Date:          20180118,
			Label:         "problem_report",
			PreviousLabel: "problem_report",
			StatusID:      "4",
		})
		if err != nil {
			panic(err)
		}

	*/
}

func addDatasets() {

	documents = append(documents, Document{
		Id:   "0",
		Text: "Text 1",
	})
	documents = append(documents, Document{
		Id:   "1",
		Text: "Text 2",
	})
	documents = append(documents, Document{
		Id:   "2",
		Text: "Text 3",
	})

	/*
	 * Insert fake datasets
	 */
	err := mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
		UploadedAt: time.Now(),
		Name:       "test_dataset_1",
		Size:       3,
		Documents:  documents,
	})
	if err != nil {
		panic(err)
	}

	err = mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
		UploadedAt: time.Now(),
		Name:       "test_dataset_2",
		Size:       3,
		Documents:  documents,
	})
	if err != nil {
		panic(err)
	}

	err = mongoClient.DB(database).C(collectionDataset).Insert(Dataset{
		UploadedAt: time.Now(),
		Name:       "test_dataset_3",
		Size:       3,
		Documents:  documents,
	})
	if err != nil {
		panic(err)
	}

}

func tearDown() {
	fmt.Println("--- --- tear down")
	mongoClient.Close()
}

type endpoint struct {
	method string
	url    string
}

func (e endpoint) withVars(vs ...interface{}) endpoint {
	e.url = fmt.Sprintf(e.url, vs...)
	return e
}

func (e endpoint) executeRequest(payload interface{}) (error, *httptest.ResponseRecorder) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(payload)
	if err != nil {
		return err, nil
	}

	req, err := http.NewRequest(e.method, e.url, body)
	if err != nil {
		return err, nil
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return nil, rr
}

func (e endpoint) mustExecuteRequest(payload interface{}) *httptest.ResponseRecorder {
	err, rr := e.executeRequest(payload)
	if err != nil {
		panic(errors.Wrap(err, `Could not execute request`))
	}

	return rr
}

func isSuccess(code int) bool {
	return code >= 200 && code < 300
}

func assertSuccess(t *testing.T, rr *httptest.ResponseRecorder) {
	if !isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected success.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}
func assertFailure(t *testing.T, rr *httptest.ResponseRecorder) {
	if isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected failure.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}

func assertJsonDecodes(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), v)
	if err != nil {
		t.Error(errors.Wrap(err, "Expected valid json array"))
	}
}

func TestPostDataset(t *testing.T) {
	// Test with normal dataset
	// Test with wrong file ending
}

func TestPostAddGroundtruth(t *testing.T) {
	// Test with normal groundtruth
	// Test with wrong dataset name
	// Test with wrong file ending
}

func TestPostDetectionResult(t *testing.T) {
	// Test with normal result
	// Test with some value missing
}

func TestPostUpdateResultName(t *testing.T) {
	// Test with normal Result
	// Test with non-existent result
}

func TestGetDataset(t *testing.T) {
	// Test normal
	// Test non-existent dataset
}

func TestGetAllDatasets(t *testing.T) {
	// Test normal
	// Test with no datasets
	ep := endpoint{"GET", "/hitec/repository/concepts/dataset/all"}

	// Test for failure
	response := ep.mustExecuteRequest(nil)
	var content []string
	assertJsonDecodes(t, response, &content)
	assert.Empty(t, content)

	addDatasets()

	// Test for success
	response = ep.mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 3)
}

func TestGetAllDetectionResults(t *testing.T) {
	// Test normal
	// Test with no results
}

func TestDeleteDataset(t *testing.T) {
	// Test normal
	// Test non-existent dataset
}

func TestDeleteResult(t *testing.T) {
	// Test with normal Result
	// Test with non-existent result
}

/*func TestPostTweet(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/store/tweet/"}
	assertFailure(t, ep.mustExecuteRequest(invalidArrayPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidTweet))
	assertSuccess(t, ep.mustExecuteRequest(tweets))
}*/

/*func TestPostClassifiedTweet(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/store/classified/tweet/"}
	assertFailure(t, ep.mustExecuteRequest(invalidArrayPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidTweet))

	tweet := tweets[0]
	tweet.TweetClass = "problem_report"
	tweet.ClassifierCertainty = 80
	tweet.Sentiment = "NEGATIVE"
	tweet.SentimentScore = -2
	assertSuccess(t, ep.mustExecuteRequest([]Tweet{tweet}))
}*/

/*func TestPostObservableTwitter(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/store/observable/"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
	}))

	// Test for success
	correctObservable := ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
		Lang:        "en",
	}
	assertSuccess(t, ep.mustExecuteRequest(correctObservable))

	MongoDeleteObservableTwitter(mongoClient, correctObservable)
}*/

/*func TestPostLabelTwitter(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/label/tweet/"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(TweetLabel{}))

	// Test for success
	tweetLabel := TweetLabel{
		Date:          20190131,
		Label:         "problem_report",
		PreviousLabel: "problem_report",
		StatusID:      "1234",
	}
	assertSuccess(t, ep.mustExecuteRequest(tweetLabel))

	err := mongoClient.DB(database).C(collectionTweetLabel).Remove(tweetLabel)
	assert.NoError(t, err, "Could not remove tweet label fro db")
}*/

/*func TestPostTweetTopics(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/store/topics"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	// Test for success
	tweet := tweets[0]
	tweet.Topics = TweetTopics{
		FirstClass: TweetClass{
			Label: "Network",
			Score: 0.3,
		},
		SecondClass: TweetClass{
			Label: "Devices",
			Score: 0.6,
		},
	}
	assertSuccess(t, ep.mustExecuteRequest(tweet))
}*/

/*func TestPostCheckAccessKey(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/access_key"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	// Test for success
	response := ep.mustExecuteRequest(existingAccessKey)
	var message ResponseMessage
	assertJsonDecodes(t, response, &message)
	assert.True(t, message.Status)

	response = ep.mustExecuteRequest(notExistingAccessKey)
	assertJsonDecodes(t, response, &message)
	assert.False(t, message.Status)
}*/

/*func TestPostUpdateAccessKeyConfiguration(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/access_key/update"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	// Test for success
	key := existingAccessKey
	key.Configuration.Topics = []string{"Contract", "Network"}
	assertSuccess(t, ep.mustExecuteRequest(key))
}*/

/*func TestGetTweetOfClass(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/%s/class/%s"}

	// Test for failure
	response := ep.withVars("should", "fail").mustExecuteRequest(nil)
	var content []Tweet
	assertJsonDecodes(t, response, &content)
	assert.Empty(t, content)

	// Test for success
	response = ep.withVars("WindItalia", "problem_report").mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 1)
}*/

/*func TestGetAllTweetsOfAccount(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/%s/all"}

	// Test for failure
	response := ep.withVars("shouldfail").mustExecuteRequest(nil)
	var content []Tweet
	assertJsonDecodes(t, response, &content)
	assert.Empty(t, content)

	// Test for success
	response = ep.withVars("Tre_It").mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 3)
}*/

/*func TestGetAllUnlabeledTweetsOfAccount(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/%s/all/unlabeled"}

	// Test for failure
	response := ep.withVars("shouldfail").mustExecuteRequest(nil)
	assertFailure(t, response)

	// Test for success
	response = ep.withVars("Tre_It").mustExecuteRequest(nil)
	var content []Tweet
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 2)
}*/

/*func TestGetAllTweetsOfAccountForCurrentWeek(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/%s/currentweek"}

	// Test for failure
	response := ep.withVars("shouldfail").mustExecuteRequest(nil)
	var content []Tweet
	assertJsonDecodes(t, response, &content)
	assert.Empty(t, content)

	// Test for success
	response = ep.withVars("Tre_It").mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 1)
}*/

/*func TestGetAllUnclassifiedTweetsOfAccount(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/%s/lang/%s/unclassified"}

	// Test for success
	response := ep.withVars("Tre_It", "it").mustExecuteRequest(nil)
	var tweets []Tweet
	assertJsonDecodes(t, response, &tweets)
	assert.Len(t, tweets, 1)
}*/

/*func TestGetAllTwitterAccountNames(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/account_name/all"}

	// Test for success
	response := ep.mustExecuteRequest(nil)
	var content TwitterAccount
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content.Names, 2)
}*/

/*func TestGetAllLabeledTweets(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/labeledtweets/all"}

	// Test for success
	response := ep.mustExecuteRequest(nil)
	var content []TweetLabel
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 1)
}*/

/*func TestGetObservablesTwitter(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/twitter/observables"}

	// Test for success
	response := ep.mustExecuteRequest(nil)
	var content []TweetLabel
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 1)
}*/

/*func TestPostAccessKeyConfiguration(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/twitter/access_key/configuration"}
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	notExistingKeyBody := map[string]string{"access_key": notExistingAccessKey.Key}
	assertFailure(t, ep.mustExecuteRequest(notExistingKeyBody))

	existingKeyBody := map[string]string{"access_key": existingAccessKey.Key}
	assertSuccess(t, ep.mustExecuteRequest(existingKeyBody))
	response := ep.mustExecuteRequest(existingKeyBody)
	assertSuccess(t, response)
	var config AccessKeyConfiguration
	assertJsonDecodes(t, response, &config)
}*/

/*func TestDeleteObservableTwitter(t *testing.T) {
	ep := endpoint{"DELETE", "/hitec/repository/twitter/observables"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	ep.mustExecuteRequest(ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
	})
	observables := MongoGetAllObservableTwitter(mongoClient)
	assert.Len(t, observables, 1)

	// Test for success
	assertSuccess(t, ep.mustExecuteRequest(ObservableTwitter{
		AccountName: "TestObserverable",
		Interval:    "2h",
		Lang:        "en",
	}))

	observables = MongoGetAllObservableTwitter(mongoClient)
	assert.Empty(t, observables)
}*/
