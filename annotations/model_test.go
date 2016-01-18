package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeURIValidatorWithValidTypeURIs(t *testing.T) {
	predicates := []string{
		"http://www.ft.com/ontology/annotation/mentions",
		"http://www.ft.com/ontology/annotation/about",
		"http://www.ft.com/ontology/annotation/isAnnotatedBy",
		"http://www.ft.com/ontology/annotation/describes",
	}
	for _, predicate := range validTypeURIs {
		err := validatePredicate(typeURI)
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
