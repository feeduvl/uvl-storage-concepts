package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"strings"
	"io/ioutil"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	contentTypeKey     = "Content-Type"
	contentTypeValJSON = "application/json"
)

var mongoClient *mgo.Session

func main() {
	mongoClient = MongoGetSession(os.Getenv("MONGO_IP"), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), database)
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
	router.HandleFunc("/hitec/repository/concepts/store/groundtruth/", postAddGroundTruth).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/detection/result/", postDetectionResult).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/detection/result/name", postUpdateResultName).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/annotation/", postAnnotation).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/agreement/", postAgreement).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/annotation/relationships/", postAllRelationshipNames).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/annotation/tores/", postAllToreTypes).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/reddit_crawler/jobs", postCrawlerJobs).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/app_review_crawler/jobs", postAppReviewCrawlerJobs).Methods("POST")
	router.HandleFunc("/hitec/repository/concepts/store/recommendations/", postRecommendations).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", getDataset).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/dataset/all", getAllDatasets).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/detection/result/all", getAllDetectionResults).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/name/{annotation}", getAnnotation).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/agreement/name/{agreement}", getAgreement).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/relationships", getAllRelationshipNames).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/tores", getAllToreTypes).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/all", getAllAnnotations).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/agreement/all", getAllAgreements).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/dataset/{dataset}", getAnnotationsForDataset).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/crawler_jobs/all", getCrawlerJobs).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/app_review_crawler_jobs/all", getAppReviewCrawlerJobs).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/recommendationTores/{codename}", getRecommendationTores).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotationcodes/all", getAllCodesFromAnnotations).Methods("GET")

	// Delete
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", deleteDataset).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/detection/result/{result}", deleteResult).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/annotation/name/{annotation}", deleteAnnotation).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/agreement/name/{agreement}", deleteAgreement).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/store/reddit_crawler/jobs/{job}", deleteCrawlerJob).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/store/app_review_crawler/jobs/{job}", deleteAppReviewCrawlerJob).Methods("DELETE")

	// Update
	router.HandleFunc("/hitec/repository/concepts/store/reddit_crawler/jobs/{job}", updateCrawlerJob).Methods("PUT")
	router.HandleFunc("/hitec/repository/concepts/store/app_review_crawler/jobs/{job}", updateAppReviewCrawlerJob).Methods("PUT")


	return router
}

func handleErrorWithRequest(err error, w http.ResponseWriter) {
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
}

//  store an existing annotation
func postAnnotation(w http.ResponseWriter, r *http.Request) {
	var annotation Annotation
	err := json.NewDecoder(r.Body).Decode(&annotation)

	fmt.Printf("postAnnotation called. Annotation: %s\n", annotation.Name)

	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertAnnotation(m, annotation)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

//  store an existing agreement
func postAgreement(w http.ResponseWriter, r *http.Request) {
	var agreement Agreement
	err := json.NewDecoder(r.Body).Decode(&agreement)

	fmt.Printf("postAgreement called. Agreement: %s\n", agreement.Name)

	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertAgreement(m, agreement)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postDataset(w http.ResponseWriter, r *http.Request) {

	var dataset Dataset
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("postDataset called. Dataset: %s\n", dataset.Name)

	// validate dataset
	err = validateDataset(dataset)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertDataset(m, dataset)
	handleErrorWithRequest(err, w)

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
		return
	}

	fmt.Printf("postDetectionResult called. Method: %s, Time: %s \n", result.Method, result.StartedAt)
	fmt.Printf("Got result: %v\n", result)


	// validate result
	err = validateResult(result)
	if err != nil {
		fmt.Printf("ERROR validating json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertResult(m, result)
	handleErrorWithRequest(err, w)

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
		return
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
	handleErrorWithRequest(err, w)

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func postAddGroundTruth(w http.ResponseWriter, r *http.Request) {

	// parse request
	var dataset Dataset
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("postAddGroundTruth called. Dataset Name: %s. \n", dataset.Name)

	// retrieve dataset
	m := mongoClient.Copy()
	defer m.Close()
	data := MongoGetDataset(m, dataset.Name)

	if data.Name != dataset.Name {
		fmt.Printf("Error adding groundtruth, dataset does not exist.\n")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if dataset.Name == "" {
		fmt.Printf("Error adding groundtruth, datset name invalid.\n")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data.GroundTruth = dataset.GroundTruth

	// insert updated result
	err = MongoInsertDataset(m, data)
	handleErrorWithRequest(err, w)

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
	_ = json.NewEncoder(w).Encode(dataset)
}

func postAllToreTypes(w http.ResponseWriter, r *http.Request) {

	fmt.Println("postAllToreTypes")
	m := mongoClient.Copy()
	defer m.Close()
	var body = bson.M{fieldToreTypes: new([]string)}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("Error decoding request: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("Got body: %v\n", body)
	var names []string
	for _, value := range body[fieldToreTypes].([]interface{}) {
		fmt.Printf("element: %v\n", value)
		names = append(names, value.(string))
	}

	err = MongoPostAllTORE(m, names)
	if err != nil {
		fmt.Printf("Error posting all tore types: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getAllToreTypes(w http.ResponseWriter, r *http.Request) {

	m := mongoClient.Copy()
	defer m.Close()
	names := MongoGetAllTORE(m)
	if names == nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		_ = json.NewEncoder(w).Encode(bson.M{"tores": names})
	}
}

func postAllRelationshipNames(w http.ResponseWriter, r *http.Request) {

	fmt.Println("postAllRelationshipNames")
	m := mongoClient.Copy()
	defer m.Close()
	var body = bson.M{fieldRelationshipNames: new([]string), "owners": new([]string)}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("Error decoding request: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("Got body: %v\n", body)
	var names []string
	var owners []string

	var bodyOwners = body["owners"].([]interface{})

	for index, value := range body[fieldRelationshipNames].([]interface{}) {
		fmt.Printf("element: %v\n", value)
		names = append(names, value.(string))
		owners = append(owners, bodyOwners[index].(string))
	}

	err = MongoPostAllRelationshipNames(m, names, owners)
	if err != nil {
		fmt.Printf("Error posting all relationship names: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getAllRelationshipNames(w http.ResponseWriter, r *http.Request) {

	m := mongoClient.Copy()
	defer m.Close()
	names, owners := MongoGetAllRelationshipNames(m)
	if names == nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		_ = json.NewEncoder(w).Encode(bson.M{"relationship_names": names, "owners": owners})
	}
}

// getAnnotation return the annotation with a given name
func getAnnotation(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	annotationName := params["annotation"]

	fmt.Println("REST call: getAnnotation, params: " + annotationName)

	// retrieve data from dataset
	m := mongoClient.Copy()
	defer m.Close()
	annotation := MongoGetAnnotation(m, annotationName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(annotation)
}

// getAgreement return the agreement with a given name
func getAgreement(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	agreementName := params["agreement"]

	fmt.Println("REST call: getAgreement, params: " + agreementName)

	// retrieve data from dataset
	m := mongoClient.Copy()
	defer m.Close()
	agreement := MongoGetAgreement(m, agreementName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(agreement)
}

// getAnnotationsForDataset return all annotations for a given dataset
func getAnnotationsForDataset(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	dataset := params["dataset"]

	fmt.Println("REST call: getAnnotationsForDataset, params: " + dataset)

	// retrieve data from dataset
	m := mongoClient.Copy()
	defer m.Close()
	annotations := MongoGetAnnotationsForDataset(m, dataset)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(annotations)
}

func getAllAnnotations(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getAllAnnotations\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	annotations := MongoGetAllAnnotations(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(annotations)

}

func getAllAgreements(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getAllAgreements\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	agreements := MongoGetAllAgreements(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(agreements)

}

func getAllDatasets(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getAllDatasets\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	datasets := MongoGetAllDatasets(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(datasets)

}

func getAllDetectionResults(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getAllDetectionResults\n")

	// retrieve all Results
	m := mongoClient.Copy()
	defer m.Close()
	results := MongoGetAllResults(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(results)

}

func deleteAnnotation(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	annotationName := params["annotation"]

	fmt.Printf("REST call: deleteAnnotation - %s\n", annotationName)

	m := mongoClient.Copy()
	defer m.Close()
	err := MongoDeleteAnnotation(m, annotationName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Annotation successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting annotation: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete annotation", Status: false})
	}
}

func deleteAgreement(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	agreementName := params["agreement"]

	fmt.Printf("REST call: deleteAgreement - %s\n", agreementName)

	m := mongoClient.Copy()
	defer m.Close()
	err := MongoDeleteAgreement(m, agreementName)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Agreement successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting agreement: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete agreement", Status: false})
	}
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
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Dataset successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete dataset", Status: false})
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
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
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
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Result successfully deleted", Status: true})
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete result", Status: false})
	}
}

func getCrawlerJobs(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getCrawlerJobs\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	crawlerJobs := MongoGetCrawlerJobs(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(crawlerJobs)

}

func postCrawlerJobs(w http.ResponseWriter, r *http.Request) {
	var crawlerJobs CrawlerJobs

	s, err := ioutil.ReadAll(r.Body) 
	if err != nil {
		panic(err) 
	}

	err = json.Unmarshal(s, &crawlerJobs)
	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err) 
	}

	m := mongoClient.Copy()
	defer m.Close()
	err = MongoInsertCrawlerJobs(m, crawlerJobs)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}

func deleteCrawlerJob(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	crawlerJobDate := params["job"]

	// error finding
	for k, v := range mux.Vars(r) {
		log.Printf("key=%v, value=%v", k, v)
	}

	fmt.Printf("REST call: deleteCrawlerJob: ")
	fmt.Printf(crawlerJobDate)

	_t := "{\"date\": \"" + crawlerJobDate + "\"}"
	var t Date
	err := json.NewDecoder(strings.NewReader(_t)).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
		fmt.Printf("ERROR parsing date: %s date: %s\n", err, crawlerJobDate)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteCrawlerJob(m, t.Date)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Crawler job successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting crawler job: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete crawler job", Status: false})
	}
}



func updateCrawlerJob(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	crawlerJobDate := params["job"]

	// error finding
	for k, v := range mux.Vars(r) {
		log.Printf("key=%v, value=%v", k, v)
	}

	fmt.Printf("REST call: updateCrawlerJob: ")
	fmt.Printf(crawlerJobDate)

	_t := "{\"date\": \"" + crawlerJobDate + "\"}"
	var t Date
	err := json.NewDecoder(strings.NewReader(_t)).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
		fmt.Printf("ERROR parsing date: %s date: %s\n", err, crawlerJobDate)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoUpdateCrawlerJob(m, t.Date)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Crawler job successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting crawler job: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete crawler job", Status: false})
	}
}

func getAppReviewCrawlerJobs(w http.ResponseWriter, _ *http.Request) {

	fmt.Printf("REST call: getCrawlerJobs\n")

	// retrieve all dataset names
	m := mongoClient.Copy()
	defer m.Close()
	crawlerJobs := MongoGetAppReviewCrawlerJobs(m)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(crawlerJobs)

}

func postAppReviewCrawlerJobs(w http.ResponseWriter, r *http.Request) {
	var appReviewCrawlerJobs AppReviewCrawlerJobs

	s, err := ioutil.ReadAll(r.Body) 
	if err != nil {
		panic(err) 
	}
	fmt.Printf("%+v\n", s)
	err = json.Unmarshal(s, &appReviewCrawlerJobs)
	fmt.Printf("%+v\n", appReviewCrawlerJobs)
	if err != nil {	
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err) 
	}
	m := mongoClient.Copy()
	defer m.Close()
	
	err = MongoInsertAppReviewCrawlerJobs(m, appReviewCrawlerJobs)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	
	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}


func deleteAppReviewCrawlerJob(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	crawlerJobDate := params["job"]

	// error finding
	for k, v := range mux.Vars(r) {
		log.Printf("key=%v, value=%v", k, v)
	}

	fmt.Printf("REST call: deleteCrawlerJob: ")
	fmt.Printf(crawlerJobDate)

	_t := "{\"date\": \"" + crawlerJobDate + "\"}"
	var t Date
	err := json.NewDecoder(strings.NewReader(_t)).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
		fmt.Printf("ERROR parsing date: %s date: %s\n", err, crawlerJobDate)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoDeleteAppReviewCrawlerJob(m, t.Date)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Crawler job successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting crawler job: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete crawler job", Status: false})
	}
}



func updateAppReviewCrawlerJob(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	crawlerJobDate := params["job"]

	// error finding
	for k, v := range mux.Vars(r) {
		log.Printf("key=%v, value=%v", k, v)
	}

	fmt.Printf("REST call: updateCrawlerJob: ")
	fmt.Printf(crawlerJobDate)

	_t := "{\"date\": \"" + crawlerJobDate + "\"}"
	var t Date
	err := json.NewDecoder(strings.NewReader(_t)).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not parse date", Status: false})
		fmt.Printf("ERROR parsing date: %s date: %s\n", err, crawlerJobDate)
		return
	}

	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoUpdateAppReviewCrawlerJob(m, t.Date)

	// write the response
	w.Header().Set(contentTypeKey, contentTypeValJSON)
	if ok == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Crawler job successfully deleted", Status: true})
		return
	} else {
		fmt.Printf("error deleting crawler job: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Message: "Could not delete crawler job", Status: false})
	}
}

func getRecommendationTores(w http.ResponseWriter, r *http.Request) {
    // get request param
	params := mux.Vars(r)
	codename := params["codename"]

	fmt.Println("REST call: getRecommendationTores, params: ", codename)

	m := mongoClient.Copy()
	defer m.Close()
	recommendation := MongoGetRecommendation(m, codename)
	recommendationTores := []string{}
	recommendationTores = append(recommendationTores, recommendation.Torecodes...)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bson.M{"recommendationTores": recommendationTores})
}

func getAllCodesFromAnnotations(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("REST call: getAllAnnotationCodes\n")

    // retrieve all dataset names
    m := mongoClient.Copy()
    defer m.Close()
    annotations := MongoGetAllAnnotationsCodes(m)

    // write the response
    w.Header().Set(contentTypeKey, contentTypeValJSON)
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(annotations)
}

// store all recommendations
func postRecommendations(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("REST call: postRecommendations\n")
	var recommendations []Recommendation
	err := json.NewDecoder(r.Body).Decode(&recommendations)

	if err != nil {
		fmt.Printf("ERROR decoding json: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()

	err = MongoDeleteRecommendationAll(m)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	err = MongoInsertManyRecommendations(m, recommendations)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeKey, contentTypeValJSON)
}