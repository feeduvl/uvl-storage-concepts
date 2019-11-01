package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testStrIn = "Leo leo Leonie 12 123 1234 12345 123456 1234567 #1234567890 nAbCdEf01G23h456I And some more non-matching @content"
var testStrOut = "[PERSON] [PERSON] Leonie 12 [PHONE] [PHONE] 12345 123456 [PHONE] [PHONE] [FISCALCODE] And some more non-matching @content"

func TestAnonymizeString(t *testing.T) {
	assert.Equal(t, AnonymizeString(testStrIn), testStrOut)
}

func TestAnonymizedString_MarshalJSON(t *testing.T) {
	type TestObject struct {
		Value AnonymizedString `json:"value"`
	}

	testObject := TestObject{Value: AnonymizedString(testStrIn)}

	jsonBytes, err := json.Marshal(testObject)
	if err != nil {
		t.Error(err)
	}

	assert.JSONEq(t, fmt.Sprintf(`{"value": "%s"}`, testStrOut), string(jsonBytes))
}
