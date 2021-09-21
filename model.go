package main

import (
	"gopkg.in/validator.v2"
	"time"
)

// The Annotation model

type RelationshipNameKey struct {
	Index * int `json:"index" bson:"index"`
	RelationshipName string `json:"relationship_name" bson:"relationship_name"`
}

type ClusterRelationship struct {
	TokenClusters []*int `json:"token_clusters" bson:"token_clusters"`
	RelationshipNames []RelationshipNameKey `json:"relationship_names" bson:"relationship_names"`
	Index *int `json:"index" bson:"index"`
}

type TokenCluster struct {
	Tokens []*int `json:"tokens" bson:"tokens"`
	Name string `json:"name" bson:"name"`
	Tore string `json:"tore" bson:"tore"`
	Index *int `json:"index" bson:"index"`
	RelationshipMemberships []*int `json:"relationship_memberships" bson:"relationship_memberships"`
}

type Token struct {
	Index *int `json:"index" bson:"index"`
	Name string `validate:"nonzero" json:"name" bson:"name"`
	Lemma string `validate:"nonzero" json:"lemma" bson:"lemma"`
	Pos string `validate:"nonzero" json:"pos" bson:"pos"`
	TokenCluster *int `json:"token_cluster" bson:"token_cluster"`
}

type Annotation struct {
	UploadedAt time.Time `validate:"nonzero" json:"uploaded_at" bson:"uploaded_at"`
	Name string `validate:"nonzero" json:"name" bson:"name"`
	Dataset    string    `validate:"nonzero" json:"dataset" bson:"dataset"`

	Tokens []Token `json:"tokens" bson:"tokens"`
	TokenClusters []TokenCluster `json:"token_clusters" bson:"token_clusters"`
	ClusterRelationships []ClusterRelationship `json:"cluster_relationships" bson:"cluster_relationships"`
}

// end Annotation model

// Dataset model
type Dataset struct {
	UploadedAt  time.Time      `validate:"nonzero" json:"uploaded_at" bson:"uploaded_at"`
	Name        string         `validate:"nonzero" json:"name" bson:"name"`
	Size        int            `json:"size" bson:"size"`
	Documents   []Document     `json:"documents" bson:"documents"`
	GroundTruth []TruthElement `json:"ground_truth" bson:"ground_truth"`
}

//TruthElement model
type TruthElement struct {
	Id    string `json:"id" bson:"id"`
	Value string `json:"value"  bson:"value"`
}

// Document model
type Document struct {
	Number int    `json:"number" bson:"number"`
	Text   string `validate:"nonzero" json:"text"  bson:"text"`
	Id     string `json:"id" bson:"id"`
}

// Result model
type Result struct {
	Method      string                 `validate:"nonzero" json:"method" bson:"method"`
	Status      string                 `validate:"nonzero" json:"status" bson:"status"`
	StartedAt   time.Time              `validate:"nonzero" json:"started_at" bson:"started_at"`
	DatasetName string                 `validate:"nonzero" json:"dataset_name" bson:"dataset_name"`
	Params      map[string]string      `json:"params" bson:"params"`
	Topics      map[string]interface{} `json:"topics" bson:"topics"`
	DocTopic    map[string]interface{} `json:"doc_topic" bson:"doc_topic"`
	Metrics     map[string]interface{} `json:"metrics" bson:"metrics"`
	Name        string                 `json:"name" bson:"name"`
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

// Date model, helper for parsing dates
type Date struct {
	Date time.Time `json:"date"`
}

func (result *Result) validate() error {
	return validator.Validate(result)
}

func (dataset *Dataset) validate() error {
	return validator.Validate(dataset)
}

func (document *Document) validate() error {
	return validator.Validate(document)
}

func validateDataset(dataset Dataset) error {

	err := dataset.validate()
	if err != nil {
		return err
	}

	for _, document := range dataset.Documents {
		err := document.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func validateResult(result Result) error {

	err := result.validate()
	if err != nil {
		return err
	}

	return nil
}
