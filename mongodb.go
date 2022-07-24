package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	database                = "concepts_data"
	collectionDataset       = "dataset"
	collectionResult        = "result"
	collectionAnnotation    = "annotation"
	collectionRelationships = "relationship"
	collectionTores         = "tores"
	collectionAgreement     = "agreement"
	collectionCrawlerJobs   = "crawler_jobs"

	fieldRelationshipNames = "relationship_names"
	fieldToreTypes         = "tores"
	fieldAnnotationName    = "name"
	fieldAgreementName     = "name"
	fieldDatasetName       = "name"
	fieldDatasetUploadedAt = "uploaded_at"
	fieldResultStartedAt   = "started_at"
	fieldResultMethodName  = "method"
	fieldCrawlerJobName    = "DatasetName"
	fieldCrawlerJobDate    = "date"
)

func panicError(err error) {
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}
}

func handleErrorInsert(err error) error {
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	} else {
		return nil
	}
}

// MongoGetSession returns a session
func MongoGetSession(mongoIP, username, password string, db string) *mgo.Session {
	info := &mgo.DialInfo{
		Addrs:    []string{mongoIP},
		Timeout:  60 * time.Second,
		Database: db,
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
	panicError(err)
	// Index
	datasetSecondIndex := mgo.Index{
		Key:        []string{fieldDatasetName, fieldDatasetUploadedAt},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	err = datasetCollection.EnsureIndex(datasetSecondIndex)
	panicError(err)
	// Index
	resultIndex := mgo.Index{
		Key:        []string{fieldResultMethodName, fieldResultStartedAt},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	resultCollection := mongoClient.DB(database).C(collectionResult)
	err = resultCollection.EnsureIndex(resultIndex)
	panicError(err)
}

// MongoInsertAnnotation returns ok if the dataset was inserted or already existed
func MongoInsertAnnotation(mongoClient *mgo.Session, annotation Annotation) error {
	annotation.LastUpdated = time.Now()
	query := bson.M{fieldAnnotationName: annotation.Name}
	update := bson.M{"$set": annotation}
	_, err := mongoClient.DB(database).C(collectionAnnotation).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	}

	return nil
}

// MongoInsertAgreement returns ok if the dataset was inserted or already existed
func MongoInsertAgreement(mongoClient *mgo.Session, agreement Agreement) error {
	agreement.LastUpdated = time.Now()
	var isCompleted = calculateIsCompleted(agreement)
	agreement.IsCompleted = isCompleted
	query := bson.M{fieldAgreementName: agreement.Name}
	update := bson.M{"$set": agreement}
	_, err := mongoClient.DB(database).C(collectionAgreement).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	}

	return nil
}

func calculateIsCompleted(agreement Agreement) bool {
	for _, codeAlternative := range agreement.CodeAlternatives {
		if codeAlternative.MergeStatus == "Pending" {
			return false
		}
	}
	return true
}

// MongoInsertDataset returns ok if the dataset was inserted or already existed
func MongoInsertDataset(mongoClient *mgo.Session, dataset Dataset) error {
	query := bson.M{fieldDatasetName: dataset.Name}
	update := bson.M{"$set": dataset}
	_, err := mongoClient.DB(database).C(collectionDataset).Upsert(query, update)

	return handleErrorInsert(err)
}

// MongoInsertResult returns ok if the result was inserted or already existed
func MongoInsertResult(mongoClient *mgo.Session, result Result) error {
	query := bson.M{fieldResultMethodName: result.Method, fieldResultStartedAt: result.StartedAt}
	update := bson.M{"$set": result}
	_, err := mongoClient.DB(database).C(collectionResult).Upsert(query, update)

	return handleErrorInsert(err)
}

// MongoDeleteAnnotation return err if there was an error
func MongoDeleteAnnotation(mongoClient *mgo.Session, annotation string) error {
	_, err := mongoClient.
		DB(database).
		C(collectionAnnotation).
		RemoveAll(bson.M{fieldAnnotationName: annotation})

	return err
}

// MongoDeleteAgreement return err if there was an error
func MongoDeleteAgreement(mongoClient *mgo.Session, agreement string) error {
	_, err := mongoClient.
		DB(database).
		C(collectionAgreement).
		RemoveAll(bson.M{fieldAgreementName: agreement})

	return err
}

// MongoDeleteDataset return ok if db entry could be deleted
func MongoDeleteDataset(mongoClient *mgo.Session, dataset string) bool {
	_, err := mongoClient.
		DB(database).
		C(collectionDataset).
		RemoveAll(bson.M{fieldDatasetName: dataset})

	return err == nil
}

func MongoPostAllTORE(mongoClient *mgo.Session, tores []string) error {
	query := bson.M{fieldToreTypes: fieldToreTypes}
	update := bson.M{"$set": bson.M{fieldToreTypes: fieldToreTypes, "names": tores}}
	_, err := mongoClient.DB(database).C(collectionTores).Upsert(query, update)

	return err
}

func MongoGetAllTORE(mongoClient *mgo.Session) []string {
	names := bson.M{"names": new([]string)}
	err := mongoClient.
		DB(database).
		C(collectionTores).Find(bson.M{fieldToreTypes: fieldToreTypes}).One(&names)

	var retnames []string
	if err != nil {
		fmt.Printf("Error getting tore types: %v\n", err)
		return retnames
	}
	for _, value := range names["names"].([]interface{}) {
		retnames = append(retnames, value.(string))
	}
	return retnames
}

func MongoPostAllRelationshipNames(mongoClient *mgo.Session, names []string, owners []string) error {
	query := bson.M{fieldRelationshipNames: fieldRelationshipNames}
	update := bson.M{"$set": bson.M{fieldRelationshipNames: fieldRelationshipNames, "names": names, "owners": owners}}
	_, err := mongoClient.DB(database).C(collectionRelationships).Upsert(query, update)

	return err
}

func MongoGetAllRelationshipNames(mongoClient *mgo.Session) ([]string, []string) {
	names := bson.M{"names": new([]string), "owners": new([]string)}
	err := mongoClient.
		DB(database).
		C(collectionRelationships).Find(bson.M{fieldRelationshipNames: fieldRelationshipNames}).One(&names)

	var retnames []string
	var retOwners []string
	if err != nil {
		fmt.Printf("Error getting relationship names: %v\n", err)
		return retnames, retOwners
	}
	var owners = names["owners"].([]interface{})

	for index, value := range names["names"].([]interface{}) {
		retnames = append(retnames, value.(string))
		retOwners = append(retOwners, owners[index].(string))
	}
	return retnames, retOwners
}

// MongoGetAnnotation returns an Annotation
func MongoGetAnnotation(mongoClient *mgo.Session, annotation string) Annotation {
	var annotationObj []Annotation
	err := mongoClient.
		DB(database).
		C(collectionAnnotation).
		Find(bson.M{fieldAnnotationName: annotation}).
		All(&annotationObj)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return annotationObj[0]
}

// MongoGetAgreement returns an Agreement
func MongoGetAgreement(mongoClient *mgo.Session, agreement string) Agreement {
	var agreementObj []Agreement
	err := mongoClient.
		DB(database).
		C(collectionAgreement).
		Find(bson.M{fieldAgreementName: agreement}).
		All(&agreementObj)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return agreementObj[0]
}

// MongoGetAnnotationsForDataset returns a list of Annotations for a dataset
func MongoGetAnnotationsForDataset(mongoClient *mgo.Session, dataset string) []Annotation {
	var annotations []Annotation
	err := mongoClient.
		DB(database).
		C(collectionAnnotation).
		Find(bson.M{fieldDatasetName: dataset}).
		All(&annotations)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return annotations
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
	panicError(err)

	// Return empty dataset if not found
	if len(dataset) == 0 {
		d := Dataset{}
		return d
	} else {
		return dataset[0]
	}
}

// MongoGetResult returns a dataset
func MongoGetResult(mongoClient *mgo.Session, startedAt time.Time) Result {
	var result []Result
	err := mongoClient.
		DB(database).
		C(collectionResult).
		Find(bson.M{fieldResultStartedAt: startedAt}).
		All(&result)
	panicError(err)

	// Return empty result if not found
	if len(result) == 0 {
		r := Result{}
		return r
	} else {
		return result[0]
	}
}

// MongoGetAllAnnotations get all annotations
func MongoGetAllAnnotations(mongoClient *mgo.Session) []Annotation {

	var annotations []Annotation

	err := mongoClient.
		DB(database).
		C(collectionAnnotation).Find(bson.M{}).Select(bson.M{"uploaded_at": 1, "last_updated": 1, "name": 1, "dataset": 1}).All(&annotations)

	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	fmt.Printf("getAllAnnotations result: %v\n", annotations)

	return annotations
}

// MongoGetAllAgreements get all agreements
func MongoGetAllAgreements(mongoClient *mgo.Session) []Agreement {

	var agreements []Agreement

	err := mongoClient.
		DB(database).
		C(collectionAgreement).Find(bson.M{}).Select(bson.M{"created_at": 1, "last_updated": 1, "name": 1, "dataset": 1, "annotation_names": 1, "is_completed": 1}).All(&agreements)

	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	fmt.Printf("getAllAgreements result: %v\n", agreements)

	return agreements
}

// MongoGetAllDatasets returns a dataset
func MongoGetAllDatasets(mongoClient *mgo.Session) []string {

	var datasetNames []string

	err := mongoClient.
		DB(database).
		C(collectionDataset).
		Find(nil).
		Distinct(fieldDatasetName, &datasetNames)

	panicError(err)

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
	panicError(err)

	return results
}


// MongoGetCrawlerJobs returns all registered crawler jobs
func MongoGetCrawlerJobs(mongoClient *mgo.Session) []CrawlerJobs {

	var crawlerJobs []CrawlerJobs

	err := mongoClient.
		DB(database).
		C(collectionCrawlerJobs).Find(bson.M{}).Select(bson.M{"subreddit_names": 1, "date": 1, "occurrence": 1, "number_posts": 1, "dataset_name": 1, "request": 1}).All(&crawlerJobs)

	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	fmt.Printf("getCrawlerJobs result: %v\n", crawlerJobs)

	return crawlerJobs

}

func MongoInsertCrawlerJobs(mongoClient *mgo.Session, crawlerJob CrawlerJobs) error {
	crawlerJob.Date = time.Now()
	query := bson.M{fieldCrawlerJobName: crawlerJob.DatasetName}
	update := bson.M{"$set": crawlerJob}

	_, err := mongoClient.DB(database).C(collectionCrawlerJobs).Upsert(query, update)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return err
	}

	return nil
}

func MongoDeleteCrawlerJob(mongoClient *mgo.Session, date time.Time) error {
	_, err := mongoClient.
		DB(database).
		C(collectionCrawlerJobs).
		RemoveAll(bson.M{fieldCrawlerJobDate: date})

	return err
}