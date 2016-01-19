package annotations

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypeURIValidatorWithValidTypeURIs(t *testing.T) {
	predicates := []string{
		"http://www.ft.com/ontology/annotation/mentions",
		"http://www.ft.com/ontology/annotation/about",
		"http://www.ft.com/ontology/annotation/isAnnotatedBy",
		"http://www.ft.com/ontology/annotation/describes",
	}
	for _, predicate := range predicates {
		err := validatePredicate(predicate)
		assertion := assert.New(t)
		assertion.Nil(err)
	}
}

func TestTypeURIValidatorWithInvalidType(t *testing.T) {
	invalidPredicate := "http://www.ft.com/ontology/annotation/mentionS"
	err := validatePredicate(invalidPredicate)
	assertion := assert.New(t)
	assertion.NotNil(err)
}

func TestUnMashallingAnnotation(t *testing.T) {
	annotation := Annotation{}
	jason := `{
                "id" : "123-123-123",
                "predicate" : "about"
                }
                `
	err := json.Unmarshal([]byte(jason), &annotation)
	if err != nil {
		panic(err)
	}
	assertion := assert.New(t)
	assertion.Nil(err)
}
