package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeURIValidatorWithValidTypeURIs(t *testing.T) {
	validTypeURIs := []string{
		"http://www.ft.com/ontology/annotation/mentions",
		"http://www.ft.com/ontology/annotation/about",
		"http://www.ft.com/ontology/annotation/isAnnotatedBy",
		"http://www.ft.com/ontology/annotation/describes",
	}
	for _, typeURI := range validTypeURIs {
		err := validateTypeURI(typeURI)
		assertion := assert.New(t)
		assertion.Nil(err)
	}
}

func TestTypeURIValidatorWithInvalidType(t *testing.T) {
	invalidTypeURI := "http://www.ft.com/ontology/annotation/mentionS"
	err := validateTypeURI(invalidTypeURI)
	assertion := assert.New(t)
	assertion.NotNil(err)
}
