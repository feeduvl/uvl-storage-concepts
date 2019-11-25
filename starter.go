package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"

	"encoding/json"
	"fmt"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	contentTypeKey     = "Content-Type"
	contentTypeValJSON = "application/json"
)

var mongoClient *mgo.Session

func main() {
	// log.SetOutput(os.Stdout)
	mongoClient = MongoGetSession(os.Getenv("MONGO_IP"), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"))
	MongoCreateCollectionIndexes(mongoClient)

	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	router := makeRouter()

	fmt.Println("ri-storage-twitter MS running")
	log.Fatal(http.ListenAndServe(":9682", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()

	// Insert
	router.HandleFunc("/hitec/repository/twitter/store/tweet/", postTweet).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/store/classified/tweet/", postClassifiedTweet).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/store/observable/", postObservableTwitter).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/label/tweet/", postLabelTwitter).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/store/topics", postTweetTopics).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/access_key", postCheckAccessKey).Methods("POST")
	router.HandleFunc("/hitec/repository/twitter/access_key/update", postUpdateAccessKeyConfiguration).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/class/{tweet_class}", getTweetOfClass).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/class/{tweet_class}/limit/{limit}", getTweetOfClassLimited).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/all", getAllTweetsOfAccount).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/all/unlabeled", getAllUnlabeledTweetsOfAccount).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/currentweek", getAllTweetsOfAccountForCurrentWeek).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/{account_name}/lang/{lang}/unclassified", getAllUnclassifiedTweetsOfAccount).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/account_name/all", getAllTwitterAccountNames).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/labeledtweets/all", getAllLabeledTweets).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/observables", getObservablesTwitter).Methods("GET")
	router.HandleFunc("/hitec/repository/twitter/access_key/configuration", postAccessKeyConfiguration).Methods("POST")

	// Delete
	router.HandleFunc("/hitec/repository/twitter/observables", deleteObservableTwitter).Methods("DELETE")

	return router
}

func postTweet(w http.ResponseWriter, r *http.Request) {
	var tweets []Tweet
	err := json.NewDecoder(r.Body).Decode(&tweets)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = validateTweets(tweets)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	MongoInsertTweets(m, tweets)

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postClassifiedTweet(w http.ResponseWriter, r *http.Request) {
	var tweets []Tweet
	err := json.NewDecoder(r.Body).Decode(&tweets)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = validateTweets(tweets)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("REST call (postClassifiedTweet): update %v tweets\n", len(tweets))

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	MongoUpdateTweetsSentimentAndClass(m, tweets)

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postObservableTwitter(w http.ResponseWriter, r *http.Request) {
	var observable ObservableTwitter
	err := json.NewDecoder(r.Body).Decode(&observable)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = observable.validate()
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("REST call (postObservableTwitter): insert observable %v\n", observable)

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoInsertObservableTwitter(m, observable)

	// send response
	if ok {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postLabelTwitter(w http.ResponseWriter, r *http.Request) {
	var tweetLabel TweetLabel
	err := json.NewDecoder(r.Body).Decode(&tweetLabel)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = tweetLabel.validate()
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("REST call (postLabelTwitter): insert tweet label %v\n", tweetLabel)

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	insertionOk := MongoInsertTweetLabel(m, tweetLabel)

	updateOk := MongoUpdateTweetClassAndAnnotation(m, tweetLabel)

	// send response
	if insertionOk && updateOk {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postTweetTopics(w http.ResponseWriter, r *http.Request) {
	var tweet Tweet
	err := json.NewDecoder(r.Body).Decode(&tweet)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("REST call (postTweetTopics)\n")

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	updateOk := MongoUpdateTweetTopics(m, tweet)

	// send response
	if updateOk {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postCheckAccessKey(w http.ResponseWriter, r *http.Request) {
	var accessKey AccessKey
	err := json.NewDecoder(r.Body).Decode(&accessKey)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("REST call (postCheckAccessKey)\n")

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	accessKeyExists := MongoGetAccessKeyExists(m, accessKey)

	// send response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Message: "access key status", Status: accessKeyExists})
}

func postUpdateAccessKeyConfiguration(w http.ResponseWriter, r *http.Request) {
	var accessKey AccessKey
	err := json.NewDecoder(r.Body).Decode(&accessKey)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("REST call (postUpdateAccessKeyConfiguration) %s\n", accessKey)

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	MongoUpdateAccessKeyConfiguration(m, accessKey)

	// send response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
}

func getTweetOfClass(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	tweetedToName := params["account_name"]
	tweetClass := params["tweet_class"]

	fmt.Println("params: ", tweetedToName, tweetClass)

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetTweetOfClass(m, tweetedToName, tweetClass)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}

func getTweetOfClassLimited(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	tweetedToName := params["account_name"]
	tweetClass := params["tweet_class"]
	limitParam := params["limit"]
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("params: ", tweetedToName, tweetClass)

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetTweetOfClassLimited(m, tweetedToName, tweetClass, limit)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}

func getAllTweetsOfAccount(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	accountName := params["account_name"]

	fmt.Printf("REST call: getAllTweetsOfAccount, account %s\n", accountName)

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetAllTweetsOfAccountName(m, accountName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}

func getAllUnclassifiedTweetsOfAccount(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	accountName := params["account_name"]
	lang := params["lang"]

	fmt.Printf("REST call: getAllUnclassifiedTweetsOfAccount, account %s and lang %s \n", accountName, lang)

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetUnclassifiedAllTweetsOfAccountName(m, accountName, lang)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}

func getAllUnlabeledTweetsOfAccount(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	accountName := params["account_name"]

	fmt.Printf("REST call: getAllUnlabeledTweetsOfAccount, account %s\n", accountName)

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetAllUnlabeledTweetsOfAccountName(m, accountName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if len(tweets) > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tweets)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "could not find any matching tweet", Status: false})
	}
}

func getAllTweetsOfAccountForCurrentWeek(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	accountName := params["account_name"]

	fmt.Printf("REST call: getAllTweetsOfAccountForCurrentWeek, account %s\n", accountName)

	from, _ := strconv.Atoi(time.Now().AddDate(0, 0, -6).Format("20060102"))
	to, _ := strconv.Atoi(time.Now().Format("20060102"))

	m := mongoClient.Copy()
	defer m.Close()
	tweets := MongoGetAllTweetsOfAccountForCurrentWeek(m, accountName, from, to)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}

func getAllTwitterAccountNames(w http.ResponseWriter, r *http.Request) {
	m := mongoClient.Copy()
	defer m.Close()

	fmt.Printf("REST call: getAllTwitterAccountNames\n")

	twitterAccounts := MongoGetAllTwitterAccounts(m)

	fmt.Printf("REST call: getAllTwitterAccountNames, result %v\n", twitterAccounts)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(twitterAccounts)
}

func getAllLabeledTweets(w http.ResponseWriter, r *http.Request) {
	m := mongoClient.Copy()
	defer m.Close()
	labeledTweets := MongoGetAllLabeledTweets(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(labeledTweets)
}

func getObservablesTwitter(w http.ResponseWriter, r *http.Request) {
	m := mongoClient.Copy()
	defer m.Close()
	observables := MongoGetAllObservableTwitter(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(observables)
}

func deleteObservableTwitter(w http.ResponseWriter, r *http.Request) {
	var observable ObservableTwitter
	err := json.NewDecoder(r.Body).Decode(&observable)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("REST call: deleteObservableTwitter for %v\n", observable)

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteObservableTwitter(m, observable)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "observable successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "could not delete observable", Status: false})
	}
}

func postAccessKeyConfiguration(w http.ResponseWriter, r *http.Request) {
	var accessKey AccessKey
	err := json.NewDecoder(r.Body).Decode(&accessKey)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()

	w.Header().Set(contentTypeKey, contentTypeValJSON)
	accessKeyExists := MongoGetAccessKeyExists(m, accessKey)
	if !accessKeyExists {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		accessKeyConfiguration := MongoGetAccessKeyConfiguration(m, accessKey)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accessKeyConfiguration)
	}
}
