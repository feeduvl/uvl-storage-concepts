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
	router.HandleFunc("/hitec/repository/concepts/store/annotation/relationships/", postAllRelationshipNames).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", getDataset).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/dataset/all", getAllDatasets).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/detection/result/all", getAllDetectionResults).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/name/{annotation}", getAnnotation).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/relationships", getAllRelationshipNames).Methods("GET")
	router.HandleFunc("/hitec/repository/concepts/annotation/all", getAllAnnotations).Methods("GET")

	// Delete
	router.HandleFunc("/hitec/repository/concepts/dataset/name/{dataset}", deleteDataset).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/detection/result/{result}", deleteResult).Methods("DELETE")
	router.HandleFunc("/hitec/repository/concepts/annotation/name/{annotation}", deleteAnnotation).Methods("DELETE")

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

func postAllRelationshipNames(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("postAllRelationshipNames")
	m := mongoClient.Copy()
	defer m.Close()
	var body = bson.M{fieldRelationshipNames: new([]string)}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("Error decoding request: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("Got body: %v\n", body)
	var names []string
	for _, value := range body[fieldRelationshipNames].([]interface{}) {
		fmt.Printf("element: %v\n", value)
		names = append(names, value.(string))
	}

	err = MongoPostAllRelationshipNames(m, names)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getAllRelationshipNames(w http.ResponseWriter, r *http.Request) {

	m := mongoClient.Copy()
	defer m.Close()
	names := MongoGetAllRelationshipNames(m)
	if names == nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		_ = json.NewEncoder(w).Encode(bson.M{"relationship_names": names})
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
