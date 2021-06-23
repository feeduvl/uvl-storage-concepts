package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"strings"

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

	fmt.Println("uvl-storage-concepts MS running")
	log.Fatal(http.ListenAndServe(":9684", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()

	// Insert
	router.HandleFunc("/hitec/repository/concepts/store/dataset/", postDataset).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/detection/result/", postDetectionResult).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/detection/result/name", postUpdateResultName).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", getDataset).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/dataset/all", getAllDatasets).Methods("GET")
	//router.HandleFunc("/hitec/repository/concepts/detection/result/", getDetectionResult).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/detection/result/all", getAllDetectionResults).Methods("GET")

	// Delete
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", deleteDataset).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/detection/result/{result}", deleteResult).Methods("DELETE")

	return router
}

func postDataset(w http.ResponseWriter, r *http.Request) {

	var dataset Dataset
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	fmt.Printf("postDataset called. Dataset: %s\n", dataset.Name)

	// validate dataset
	err = validateDataset(dataset)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertDataset(m, dataset)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postDetectionResult(w http.ResponseWriter, r *http.Request) {

	// parse request
	var result Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	fmt.Printf("postDetectionResult called. Method: %s, Time: %s \n", result.Method, result.StartedAt)

	// validate result
	err = validateResult(result)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertResult(m, result)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postUpdateResultName(w http.ResponseWriter, r *http.Request) {

	// parse request
	var result Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	fmt.Printf("postUpdateResultName called. Name: %s, Time: %s \n", result.Name, result.StartedAt)

	// retrieve result
	m := mongoClient.Copy()
	defer m.Close()
	res := MongoGetResult(m, result.StartedAt)

	if res.Status != "finished" && res.Status != "failed" {
		fmt.Printf("ERROR: can not change name for result with status: %s\n", res.Status)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Name = result.Name

	// insert updated result
	err = MongoInsertResult(m, res)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func getDataset(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	datasetName := params["dataset"]

	fmt.Println("REST call: getDataset, params: ", datasetName)

	// retrieve data from dataset
	m := mongoClient.Copy()
	defer m.Close()
	dataset := MongoGetDataset(m, datasetName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataset)
}

func getAllDatasets(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("REST call: getAllDatasets\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	datasets := MongoGetAllDatasets(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)

}

func getDetectionResult(w http.ResponseWriter, r *http.Request) {

	//

}

func getAllDetectionResults(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("REST call: getAllDetectionResults\n")

	// retrieve all Results
	m := mongoClient.Copy()
	defer m.Close()
	results := MongoGetAllResults(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)

}

func deleteDataset(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	dataset := params["dataset"]

	fmt.Printf("REST call: deleteDataset - %s\n", dataset)

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteDataset(m, dataset)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "Dataset successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete dataset", Status: false})
	}
}

func deleteResult(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	result := params["result"]

	fmt.Printf("REST call: deleteResult - %s\n", result)

	_t := "{\"date\": \"" + result + "\"}"
	// parse time
	var t Date
	err := json.NewDecoder(strings.NewReader(_t)).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
		fmt.Printf("ERROR parsing date: %s date: %s\n", err, result)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteResult(m, t.Date)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "Result successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete result", Status: false})
	}
}

/*
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
*/
