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

	fmt.Println("uvl-storage-concepts MS running")
	log.Fatal(http.ListenAndServe(":9682", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()

	// Insert
	router.HandleFunc("/hitec/repository/concepts/store/dataset/", postDataset).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/detection/result/", postDetectionResult).Methods("POST")
	//router.HandleFunc("/hitec/repository/twitter/access_key", postCheckAccessKey).Methods("POST")
	//router.HandleFunc("/hitec/repository/twitter/access_key/update", postUpdateAccessKeyConfiguration).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/concepts/dataset/{dataset}", getDataset).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/dataset/all", getAllDatasets).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/detection/result/{result}", getDetectionResult).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/detection/result/all", getAllDetectionResults).Methods("GET")
	//router.HandleFunc("/hitec/repository/twitter/access_key/configuration", postAccessKeyConfiguration).Methods("POST")

	// Delete
	router.HandleFunc("/hitec/repository/concepts/dataset", deleteDataset).Methods("DELETE")

	return router
}

func postDataset(w http.ResponseWriter, r *http.Request) {
	var dataset Dataset
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate dataset
	/*err = validateTweets(tweets)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	MongoInsertDataset(m, dataset)

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postDetectionResult() {
	//
}

func getDataset() {
	//
}

func getAllDatasets() {
	//
}

func getDetectionResult() {
	//
}

func getAllDetectionResults() {
	//
}

func deleteDataset() {
	var dataset Dataset
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("REST call: deleteDataset - %v\n", dataset)

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteDataset(m, dataset)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "dataset successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Message: "could not delete dataset", Status: false})
	}
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
