package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/dbtest"
	"io/ioutil"
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
var ti = time.Now()
var invalidObjectPayload []byte
var invalidPayloadString = "payload"

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
}

func setupRouter() {
	router = makeRouter()
}

func setupDB() {
	tempDir, _ := ioutil.TempDir("", "testing")
	mockDBServer.SetPath(tempDir)

	mongoClient = mockDBServer.Session()
	MongoCreateCollectionIndexes(mongoClient)
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
	//mongoClient.Close()
	mockDBServer.Stop()
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
	ep := endpoint{"POST", "/hitec/repository/concepts/store/dataset/"}

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

	validDatasetPayload := Dataset{
		UploadedAt: time.Now(),
		Name:       "test_dataset_5",
		Size:       3,
		Documents:  documents,
	}

	// Test with normal dataset
	assertSuccess(t, ep.mustExecuteRequest(validDatasetPayload))

	d := MongoGetAllDatasets(mongoClient)
	assert.Len(t, d, 1)

	// Test with exising dataset name
	assertSuccess(t, ep.mustExecuteRequest(validDatasetPayload))

	d = MongoGetAllDatasets(mongoClient)
	assert.Len(t, d, 1)

	MongoDeleteDataset(mongoClient, "test_dataset_5")

	// Test invalid payload
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidPayloadString))

	// Test dataset with invalid document
	documents = append(documents, Document{
		Id:   "",
		Text: "",
	})

	invalidDatasetPayload := Dataset{
		UploadedAt: time.Now(),
		Name:       "test_dataset_5",
		Size:       4,
		Documents:  documents,
	}
	assertFailure(t, ep.mustExecuteRequest(invalidDatasetPayload))
}

func TestPostDetectionResult(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/concepts/store/detection/result/"}

	// Test with normal result
	res := Result{
		Method:      "lda",
		Status:      "finished",
		StartedAt:   ti,
		DatasetName: "test_dataset_2",
		Name:        "test_result",
	}
	assertSuccess(t, ep.mustExecuteRequest(res))

	resFail := Result{
		Method:      "",
		Status:      "",
		DatasetName: "test",
		Name:        "test_result_fail",
	}
	// Test with some value missing
	assertFailure(t, ep.mustExecuteRequest(resFail))

	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidPayloadString))

}

func TestPostUpdateResultName(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/concepts/store/detection/result/name"}
	// Test with normal Result
	res := Result{
		StartedAt: ti,
		Name:      "new_name",
	}

	assertSuccess(t, ep.mustExecuteRequest(res))

	// Test with non-existent result
	resFail := Result{
		StartedAt: time.Now(),
		Name:      "new_name",
	}
	assertFailure(t, ep.mustExecuteRequest(resFail))

	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidPayloadString))

}

func TestGetAllDatasets(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/concepts/dataset/all"}

	// Test with no datasets
	response := ep.mustExecuteRequest(nil)
	var content []string
	assertJsonDecodes(t, response, &content)
	assert.Empty(t, content)

	addDatasets()

	// Test normal
	response = ep.mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 3)
}

func TestGetDataset(t *testing.T) {

	// Test normal
	ep := endpoint{"GET", "/hitec/repository/concepts/dataset/name/test_dataset_2"}
	response := ep.mustExecuteRequest(nil)
	var content Dataset
	assertJsonDecodes(t, response, &content)
	assert.Equal(t, content.Name, "test_dataset_2")
	assert.Len(t, content.Documents, 7)

	// Test non-existent dataset
	ep = endpoint{"GET", "/hitec/repository/concepts/dataset/name/test_dataset_4"}
	response = ep.mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.NotEqual(t, content.Name, "test_dataset_4")

}

func TestPostAddGroundtruth(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/concepts/store/groundtruth/"}
	// Test with normal groundtruth
	var gt []TruthElement
	d := Dataset{
		Name:        "test_dataset_2",
		GroundTruth: gt,
	}
	assertSuccess(t, ep.mustExecuteRequest(d))

	d.Name = "test_dataset_99"
	// Test with wrong dataset name
	assertFailure(t, ep.mustExecuteRequest(d))

	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))
	assertFailure(t, ep.mustExecuteRequest(invalidPayloadString))
}

func TestGetAllDetectionResults(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/concepts/detection/result/all"}
	// Test normal
	response := ep.mustExecuteRequest(nil)
	var content []Result
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 1)

	MongoDeleteResult(mongoClient, ti)
	// Test with no results
	response = ep.mustExecuteRequest(nil)
	assertJsonDecodes(t, response, &content)
	assert.Len(t, content, 0)
}

func TestDeleteDataset(t *testing.T) {
	// Test normal
	ep := endpoint{"DELETE", "/hitec/repository/concepts/dataset/name/test_dataset_1"}
	ep.mustExecuteRequest(nil)
	datasets := MongoGetAllDatasets(mongoClient)
	assert.Len(t, datasets, 2)

	// Test non-existent dataset
	ep = endpoint{"DELETE", "/hitec/repository/concepts/dataset/name/test_dataset_4"}
	ep.mustExecuteRequest(nil)
	datasets = MongoGetAllDatasets(mongoClient)
	assert.Len(t, datasets, 2)

}

func TestDeleteResult(t *testing.T) {

	res := Result{
		Method:      "seanmf",
		Status:      "finished",
		StartedAt:   ti.Truncate(time.Millisecond),
		DatasetName: "test_dataset_2",
		Name:        "test_result",
	}
	_ = MongoInsertResult(mongoClient, res)

	// Test with non-existent result
	tm := time.Now().Format("2006-01-02T15:04:05Z07:00")
	ep := endpoint{"DELETE", "/hitec/repository/concepts/detection/result/" + tm}
	assertSuccess(t, ep.mustExecuteRequest(nil))

	results := MongoGetAllResults(mongoClient)
	assert.Len(t, results, 1)

	// Test with wrong date format
	tm = ti.Format("2006-ZZ01-02T15:04:05.000Z07:00")
	fmt.Println(tm)
	ep = endpoint{"DELETE", "/hitec/repository/concepts/detection/result/" + tm}
	assertFailure(t, ep.mustExecuteRequest(nil))

	// Test with normal Result
	tm = ti.Format("2006-01-02T15:04:05.000Z07:00")
	fmt.Println(tm)
	ep = endpoint{"DELETE", "/hitec/repository/concepts/detection/result/" + tm}
	assertSuccess(t, ep.mustExecuteRequest(nil))
	results = MongoGetAllResults(mongoClient)
	assert.Len(t, results, 0)
}

func TestQueries(t *testing.T) {
	mongoClient.Close()
	assert.Panics(t, func() {
		MongoCreateCollectionIndexes(mongoClient)
	})
}

func TestErrorHandler(t *testing.T) {
	assert.Panics(t, func() {
		panicError(errors.New("Error"))
	})
	assert.NotPanics(t, func() {
		panicError(nil)
	})
	err := handleErrorInsert(errors.New("Error"))
	assert.Error(t, err)
}
