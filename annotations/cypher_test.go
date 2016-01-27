package annotations

import (
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var annotationsDriver CypherDriver

const (
	contentUUID    = "http://api.ft.com/things/32b089d2-2aae-403d-be6e-877404f586cf"
	conceptUUID    = "http://api.ft.com/things/c834adfa-10c9-4748-8a21-c08537172706"
	oldConceptUUID = "http://api.ft.com/things/ad28ddc7-4743-4ed3-9fad-5012b61fb919"
)

func TestDeleteRemovesAnnotations(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToDelete := Annotations{Annotation{
		Thing: Thing{ID: conceptUUID,
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToDelete), "Failed to write annotation")

	found, err := annotationsDriver.DeleteAll(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotation for content uuid %, conceptUUID %s", contentUUID, conceptUUID)

	annotations, found, err := annotationsDriver.Read(contentUUID)

	assert.Equal(Annotations{}, annotations, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.False(found, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.NoError(err, "Error trying to find annotation for content %s", contentUUID)
}

func TestDeleteDoesNotRemoveExistingOrganisation(t *testing.T) {

}

func TestDeleteDoesNotRemoveExistingContent(t *testing.T) {

}

func TestWriteAllValuesPresent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: conceptUUID,
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToWrite)

	cleanUp(t, contentUUID)
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: conceptUUID},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToWrite)

	cleanUp(t, contentUUID)
}

func TestUpdateWillRemovePreviousAnnotations(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	oldAnnotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: oldConceptUUID,
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, oldAnnotationsToWrite), "Failed to write annotations")
	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, oldAnnotationsToWrite)

	updatedAnnotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: conceptUUID,
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, updatedAnnotationsToWrite), "Failed to write updated annotations")
	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, updatedAnnotationsToWrite)

	cleanUp(t, contentUUID)
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getCypherDriver(t)
	err := annotationsDriver.CheckConnectivity()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func getCypherDriver(t *testing.T) CypherDriver {
	assert := assert.New(t)
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return NewCypherDriver(neoutils.StringerDb{db}, db)
}

func readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t *testing.T, contentUUID string, expectedAnnotations []Annotation) {
	assert := assert.New(t)
	storedAnnotations, found, err := annotationsDriver.Read(contentUUID)

	assert.NoError(err, "Error finding annotations for contentUUID %s", contentUUID)
	assert.True(found, "Didn't find annotations for contentUUID %s", contentUUID)
	expectedAnnotation := expectedAnnotations[0]
	storedAnnotation := storedAnnotations[0]
	assert.EqualValues(expectedAnnotation.Provenances, storedAnnotation.Provenances, "Provenances not the same")

	// In annotations write, we don't store anything other than ID for the concept (so type will only be 'Thing' and pref label will not
	// be present UNLESS the concept has been written by some other system)
	assert.Equal(expectedAnnotation.Thing.ID, storedAnnotation.Thing.ID, "Thing ID not the same")
}

func cleanUp(t *testing.T, contentUUID string) {
	assert := assert.New(t)
	found, err := annotationsDriver.DeleteAll(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %", contentUUID)
	assert.NoError(err, "Error deleting annotations for content uuid %s", contentUUID)
}
