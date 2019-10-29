package annotations

import (
	"fmt"
	"testing"

	"github.com/Financial-Times/go-logger"
	"github.com/stretchr/testify/assert"
)

const (
	contentUUID            = "32b089d2-2aae-403d-be6e-877404f586cf"
	v2PlatformVersion      = "v2"
	pacPlatformVersion     = "pac"
	v2AnnotationLifecycle  = "annotations-v2"
	pacAnnotationLifecycle = "annotations-pac"
)

func TestCreateAnnotationQuery(t *testing.T) {
	assert := assert.New(t)
	logger.InitDefaultLogger("annotations-rw")
	annotationToWrite := exampleConcept(oldConceptUUID)

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2PlatformVersion, v2AnnotationLifecycle)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	params := query.Parameters["annProps"].(map[string]interface{})
	assert.Equal(v2PlatformVersion, params["platformVersion"], fmt.Sprintf("\nExpected: %s\nActual: %s", v2PlatformVersion, params["platformVersion"]))
}

func TestGetRelationshipFromPredicate(t *testing.T) {
	var tests = []struct {
		predicate    string
		relationship string
	}{
		{"mentions", "MENTIONS"},
		{"isClassifiedBy", "IS_CLASSIFIED_BY"},
		{"", "MENTIONS"},
		{"about", "ABOUT"},
		{"hasAuthor", "HAS_AUTHOR"},
		{"hasBrand", "HAS_BRAND"},
	}

	for _, test := range tests {
		actualRelationship, err := getRelationshipFromPredicate(test.predicate)
		assert.NoError(t, err)

		if test.relationship != actualRelationship {
			t.Errorf("\nExpected: %s\nActual: %s", test.relationship, actualRelationship)
		}
	}
}

func TestCreateAnnotationQueryWithPredicate(t *testing.T) {
	assert := assert.New(t)
	logger.InitDefaultLogger("annotations-rw")
	annotationToWrite := conceptWithPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2AnnotationLifecycle, v2PlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "IS_CLASSIFIED_BY", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "IS_CLASSIFIED_BY should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithAboutPredicate(t *testing.T) {
	assert := assert.New(t)
	logger.InitDefaultLogger("annotations-rw")
	annotationToWrite := conceptWithAboutPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2AnnotationLifecycle, v2PlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "ABOUT", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "ABOUT should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithHasAuthorPredicate(t *testing.T) {
	assert := assert.New(t)
	logger.InitDefaultLogger("annotations-rw")
	annotationToWrite := conceptWithHasAuthorPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2AnnotationLifecycle, v2PlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "HAS_AUTHOR", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "HAS_AUTHOR should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithHasContributorPredicate(t *testing.T) {
	assert := assert.New(t)
	annotationToWrite := conceptWithHasContributorPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, pacAnnotationLifecycle, pacPlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "HAS_CONTRIBUTOR", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "HAS_CONTRIBUTOR should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithHasDisplayTagPredicate(t *testing.T) {
	assert := assert.New(t)
	annotationToWrite := conceptWithHasDisplayTagPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, pacAnnotationLifecycle, pacPlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "HAS_DISPLAY_TAG", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "HAS_DISPLAY_TAG should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithImplicitlyClassifiedByPredicate(t *testing.T) {
	assert := assert.New(t)
	annotationToWrite := conceptWithImplicitlyClassifiedByPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, pacAnnotationLifecycle, pacPlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "IMPLICITLY_CLASSIFIED_BY", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "IMPLICITLY_CLASSIFIED_BY should be inserted instead of MENTIONS")
}

func TestCreateAnnotationQueryWithHasBrandPredicate(t *testing.T) {
	assert := assert.New(t)
	logger.InitDefaultLogger("annotations-rw")
	annotationToWrite := conceptWithHasBrandPredicate

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, pacAnnotationLifecycle, pacPlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "HAS_BRAND", "Relationship name is not inserted!")
	assert.NotContains(query.Statement, "MENTIONS", "HAS_BRAND should be inserted instead of MENTIONS")
}
