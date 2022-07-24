package main

import (
	"gopkg.in/validator.v2"
	"time"
)

// The Annotation model

type DocWrapper struct {
	Name       string `json:"name" bson:"name"`
	BeginIndex *int   `json:"begin_index" bson:"begin_index"`
	EndIndex   *int   `json:"end_index" bson:"end_index"`
}

type TORERelationship struct {
	TOREEntity       *int   `json:"TOREEntity" bson:"TOREEntity"`
	TargetTokens     []*int `json:"target_tokens" bson:"target_tokens"`
	RelationshipName string `json:"relationship_name" bson:"relationship_name"`
	Index            *int   `json:"index" bson:"index"`
}

type Code struct {
	Tokens                  []*int `json:"tokens" bson:"tokens"`
	Name                    string `json:"name" bson:"name"`
	Tore                    string `json:"tore" bson:"tore"`
	Index                   *int   `json:"index" bson:"index"`
	RelationshipMemberships []*int `json:"relationship_memberships" bson:"relationship_memberships"`
}

type Token struct {
	Index        *int   `json:"index" bson:"index"`
	Name         string `validate:"nonzero" json:"name" bson:"name"`
	Lemma        string `validate:"nonzero" json:"lemma" bson:"lemma"`
	Pos          string `validate:"nonzero" json:"pos" bson:"pos"`
	NumNameCodes int    `json:"num_name_codes" bson:"num_name_codes"`
	NumToreCodes int    `json:"num_tore_codes" bson:"num_tore_codes"`
}

type Annotation struct {
	UploadedAt  time.Time `validate:"nonzero" json:"uploaded_at" bson:"uploaded_at"`
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`

	Name    string `validate:"nonzero" json:"name" bson:"name"`
	Dataset string `validate:"nonzero" json:"dataset" bson:"dataset"`

	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	Codes             []Code             `json:"codes" bson:"codes"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`
}

// end Annotation model
// The Agreement model

// AgreementStatistics model, the initial and current kappas. Name is unique
type AgreementStatistics struct {
	KappaName    string  `validate:"nonzero" json:"kappa_name" bson:"kappa_name"`
	InitialKappa float64 `json:"initial_kappa" bson:"initial_kappa"`
	CurrentKappa float64 `json:"current_kappa" bson:"current_kappa"`
}

// CodeAlternatives model, shows all code alternatives from all annotations, MergeStatus can be set to Pending, Accepted or Declined
type CodeAlternatives struct {
	AnnotationName string `json:"annotation_name" bson:"annotation_name"`
	MergeStatus    string `validate:"nonzero" json:"merge_status" bson:"merge_status"`
	Index          int    `json:"index" bson:"index"`

	Code Code `json:"code" bson:"code"`
}

// Agreement model
type Agreement struct {
	CreatedAt   time.Time `validate:"nonzero" json:"created_at" bson:"created_at"`
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`

	Name        string   `validate:"nonzero" json:"name" bson:"name"`
	Dataset     string   `validate:"nonzero" json:"dataset" bson:"dataset"`
	Annotations []string `json:"annotation_names" bson:"annotation_names"`

	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`

	CodeAlternatives    []CodeAlternatives    `json:"code_alternatives" bson:"code_alternatives"`
	AgreementStatistics []AgreementStatistics `json:"agreement_statistics" bson:"agreement_statistics"`

	IsCompleted bool `json:"is_completed" bson:"is_completed"`
}

// end Agreement model

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
	Codes       []Code   			   `json:"codes" bson:"codes"`
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

type CrawlerRequest struct {
	Subreddits        []string `json:"subreddits" bson:"subreddits"`
	blacklistComments []string `json:"blacklist_comments" bson:"blacklist_comments"`
	blacklistPosts    []string `json:"blacklist_posts" bson:"blacklist_posts"`
	commentDepth      int      `json:"comment_depth" bson:"comment_depth"`
	datasetName       string   `json:"dataset_name" bson:"dataset_name"`
	dateFrom          string   `json:"date_from" bson:"date_from"`
	dateTo            string   `json:"date_to" bson:"date_to"`
	minLengthComments int      `json:"min_length_comments" bson:"min_length_comments"`
	minLengthPosts    int      `json:"min_length_posts" bson:"min_length_posts"`
	newLimit          int      `json:"new_limit" bson:"new_limit"`
	postSelection     string   `json:"post_selection" bson:"post_selection"`
	replaceEmojis     bool     `json:"replace_emojis" bson:"replace_emojis"`
	replaceUrls       bool     `json:"replace_urls" bson:"replace_urls"`
}

// Crawler Jobs model
type CrawlerJobs struct {
	SubredditName string    `validate:"nonzero" json:"subreddit_names" bson:"subreddit_names"`
	Date          time.Time `validate:"nonzero" json:"date" bson:"date"`
	Occurrence    int       `json:"occurrence" bson:"occurrence"`
	NumberPosts   int       `json:"number_posts" bson:"number_posts"`
	DatasetName   string    `validate:"nonzero" json:"dataset_name" bson:"dataset_name"`
	Request       CrawlerRequest `validate:"nonzero" json:"request" bson:"request"`
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
