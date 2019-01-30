package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/dbtest"
)

var router *mux.Router
var mockDBServer dbtest.DBServer
var tweets []interface{}

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
	router = mux.NewRouter()
	// Insert
	router.HandleFunc("/hitec/repository/twitter/store/tweet/", postTweet).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/store/observable/", postObservableTwitter).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/label/tweet/", postLabelTwitter).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/class/{tweet_class}", getTweetOfClass).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/all", getAllTweetsOfAccount).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/all/unlabeled", getAllUnlabeledTweetsOfAccount).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/currentweek", getAllTweetsOfAccountForCurrentWeek).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/all", getAllTwitterAccountNames).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/labeledtweets/all", getAllLabeledTweets).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/observables", getObservablesTwitter).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/observables", deleteObservableTwitter).Methods("DELETE")
}

func setupDB() {
	tempDir, _ := ioutil.TempDir("", "testing")
	mockDBServer.SetPath(tempDir)

	mongoClient = mockDBServer.Session()
	MongoCreateCollectionIndexes(mongoClient)
}

func fillDB() {
	/*
	 * Insert fake tweets
	 */
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
		UserName:            "nytwitt",
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
		UserName:            "nytwitt",
		InReplyToScreenName: "Tre_It",
		Hashtags:            []string{"bellapersona"},
		Lang:                "it",
		Sentiment:           "NEUTRAL",
		SentimentScore:      0,
		TweetClass:          "inquiry",
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
		UserName:            "nytwitt",
		InReplyToScreenName: "Tre_It",
		Hashtags:            []string{"bellapersona"},
		Lang:                "it",
		Sentiment:           "NEUTRAL",
		SentimentScore:      0,
		TweetClass:          "inquiry",
		ClassifierCertainty: 0,
	})

	tweetBulk := mongoClient.DB(database).C(collectionTweet).Bulk()
	tweetBulk.Insert(tweets...)
	_, err := tweetBulk.Run()
	if err != nil {
		panic(err)
	}

	/*
	 * Insert fake observables
	 */
	err = mongoClient.DB(database).C(collectionObservableTwitter).Insert(ObservableTwitter{
		AccountName: "TestObserverable",
		Interval:    "2h",
		Lang:        "en",
	})
	if err != nil {
		panic(err)
	}

	/*
	 * Insert fake labels
	 */
	err = mongoClient.DB(database).C(collectionTweetLabel).Insert(TweetLabel{
		Date:          20180118,
		Label:         "problem_report",
		PreviousLabel: "problem_report",
		StatusID:      "4",
	})
	if err != nil {
		panic(err)
	}
}

func tearDown() {
	fmt.Println("--- --- tear down")
	mongoClient.Close()
	mockDBServer.Stop() // Stop shuts down the temporary server and removes data on disk.
}

func buildRequest(method, endpoint string, payload io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, endpoint, payload)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	return req
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func TestPostTweet(t *testing.T) {
	fmt.Println("start test TestPostTweet")
	var method = "POST"
	var endpoint = "/hitec/repository/twitter/store/tweet/"

	/*
	 * test for faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode([]byte(`[{ "wrong_json_format": true }]`))
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}

	req := buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	payload = new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(tweets)
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestPostObservableTwitter(t *testing.T) {
	fmt.Println("start test TestPostObservableTwitter")
	var method = "POST"
	var endpoint = "/hitec/repository/twitter/store/observable/"

	/*
	 * test for faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
	})
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req := buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	payload = new(bytes.Buffer)
	correctlyFormattedObservable := ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
		Lang:        "en",
	}
	err = json.NewEncoder(payload).Encode(correctlyFormattedObservable)
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	MongoDeleteObservableTwitter(mongoClient, correctlyFormattedObservable)
}

func TestPostLabelTwitter(t *testing.T) {
	fmt.Println("start test TestPostLabelTwitter")
	var method = "POST"
	var endpoint = "/hitec/repository/twitter/label/tweet/"

	/*
	 * test for faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(TweetLabel{})
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req := buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	payload = new(bytes.Buffer)
	tweetLabel := TweetLabel{
		Date:          20190131,
		Label:         "problem_report",
		PreviousLabel: "problem_report",
		StatusID:      "1234",
	}
	err = json.NewEncoder(payload).Encode(tweetLabel)
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	err = mongoClient.DB(database).C(collectionTweetLabel).Remove(tweetLabel)
	if err != nil {
		t.Errorf("Could not remove tweet label fro db")
	}
}
func TestGetTweetOfClass(t *testing.T) {
	fmt.Println("start test TestGetTweetOfClass")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/account_name/%s/class/%s"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "should", "fail")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsResponse []Tweet
	err := json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 0 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 0, len(tweetsResponse))
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "WindItalia", "problem_report")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	err = json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(tweetsResponse))
	}
}

func TestGetAllTweetsOfAccount(t *testing.T) {
	fmt.Println("start test TestGetAllTweetsOfAccount")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/account_name/%s/all"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "shouldfail")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsResponse []Tweet
	err := json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 0 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 0, len(tweetsResponse))
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "Tre_It")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	err = json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 3 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 3, len(tweetsResponse))
	}
}

func TestGetAllUnlabeledTweetsOfAccount(t *testing.T) {
	fmt.Println("start test TestGetAllUnlabeledTweetsOfAccount")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/account_name/%s/all/unlabeled"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "shouldfail")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "Tre_It")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsResponse []Tweet
	err := json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 2 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 2, len(tweetsResponse))
	}
}

func TestGetAllTweetsOfAccountForCurrentWeek(t *testing.T) {
	fmt.Println("start test TestGetAllTweetsOfAccount")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/account_name/%s/currentweek"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "shouldfail")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsResponse []Tweet
	err := json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 0 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 0, len(tweetsResponse))
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "Tre_It")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	err = json.NewDecoder(rr.Body).Decode(&tweetsResponse)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsResponse) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(tweetsResponse))
	}
}

func TestGetAllTwitterAccountNames(t *testing.T) {
	fmt.Println("start test TestGetAllTwitterAccountNames")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/account_name/all"
	/*
	 * test for success
	 */
	req := buildRequest(method, endpoint, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var TwitterAccount TwitterAccount
	err := json.NewDecoder(rr.Body).Decode(&TwitterAccount)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(TwitterAccount.Names) != 2 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 2, len(TwitterAccount.Names))
	}
}

func TestGetAllLabeledTweets(t *testing.T) {
	fmt.Println("start test TestGetAllLabeledTweets")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/labeledtweets/all"

	/*
	 * test for faillure
	 */
	req := buildRequest(method, endpoint, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsLabeled []TweetLabel
	err := json.NewDecoder(rr.Body).Decode(&tweetsLabeled)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsLabeled) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(tweetsLabeled))
	}
}

func TestGetObservablesTwitter(t *testing.T) {
	fmt.Println("start test TestGetObservablesTwitter")
	var method = "GET"
	var endpoint = "/hitec/repository/twitter/observables"

	/*
	 * test for Success
	 */
	req := buildRequest(method, endpoint, nil, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var tweetsLabeled []TweetLabel
	err := json.NewDecoder(rr.Body).Decode(&tweetsLabeled)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(tweetsLabeled) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(tweetsLabeled))
	}
}

func TestDeleteObservablesTwitter(t *testing.T) {
	fmt.Println("start test TestDeleteObservablesTwitter")
	var method = "DELETE"
	var endpoint = "/hitec/repository/twitter/observables"

	/*
	 * test for Faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(ObservableTwitter{
		AccountName: "Test",
		Interval:    "2h",
	})
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req := buildRequest(method, endpoint, payload, t)
	_ = executeRequest(req)

	observables := MongoGetAllObservableTwitter(mongoClient)
	if len(observables) != 1 {
		t.Errorf("Number of observables differ. Expected %d .\n Got %d instead", len(observables), 1)
	}

	/*
	 * test for Success
	 */
	payload = new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(ObservableTwitter{
		AccountName: "TestObserverable",
		Interval:    "2h",
		Lang:        "en",
	})
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	observables = MongoGetAllObservableTwitter(mongoClient)
	if len(observables) != 0 {
		t.Errorf("Number of observables differ. Expected %d .\n Got %d instead", len(observables), 0)
	}
}
