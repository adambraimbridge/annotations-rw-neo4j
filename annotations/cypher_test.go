package annotations

import (
	"fmt"
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var annotationsDriver CypherDriver

const (
	contentUUID    = "32b089d2-2aae-403d-be6e-877404f586cf"
	conceptUUID    = "c834adfa-10c9-4748-8a21-c08537172706"
	oldConceptUUID = "ad28ddc7-4743-4ed3-9fad-5012b61fb919"
)

func getURI(uuid string) string {
	return fmt.Sprintf("http://api.ft.com/things/%s", uuid)
}

func TestDeleteRemovesAnnotationsButNotConceptsOrContent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToDelete := Annotations{Annotation{
		Thing: Thing{ID: getURI(conceptUUID),
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

	checkNodeIsStillPresent(contentUUID, t)
	checkNodeIsStillPresent(conceptUUID, t)

	err = deleteNode(annotationsDriver, contentUUID)
	assert.NoError(err, "Error trying to delete content node with uuid %s", contentUUID)
	err = deleteNode(annotationsDriver, conceptUUID)
	assert.NoError(err, "Error trying to delete concept node with uuid %s", conceptUUID)

}

func TestWriteAllValuesPresent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: getURI(conceptUUID),
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

	cleanUp(t, contentUUID, []string{conceptUUID})
}

func TestWriteOnlyMandatoryValuesPresent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	annotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: getURI(conceptUUID)},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToWrite)

	cleanUp(t, contentUUID, []string{conceptUUID})
}

func TestWriteFailsForMoreThanOneProvenance(t *testing.T) {

}

func TestUpdateWillRemovePreviousAnnotations(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getCypherDriver(t)

	oldAnnotationsToWrite := Annotations{Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
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
		Thing: Thing{ID: getURI(conceptUUID),
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

	cleanUp(t, contentUUID, []string{conceptUUID})
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

func checkNodeIsStillPresent(uuid string, t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getCypherDriver(t)
	results := []struct {
		UUID string `json:"uuid"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Thing {uuid:{uuid}}) return n.uuid
		as uuid`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	err := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	assert.NoError(err, "UnexpectedError")
	assert.True(len(results) == 1, "Didn't find a node")
	assert.Equal(uuid, results[0].UUID, "Did not find correct node")
}

func cleanUp(t *testing.T, contentUUID string, conceptUUIDs []string) {
	assert := assert.New(t)
	found, err := annotationsDriver.DeleteAll(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotations for content uuid %s", contentUUID)

	err = deleteNode(annotationsDriver, contentUUID)
	assert.NoError(err, "Could not delete content node")
	for _, conceptUUID := range conceptUUIDs {
		err = deleteNode(annotationsDriver, conceptUUID)
		assert.NoError(err, "Could not delete concept node")
	}
}

func deleteNode(annotationsDriver CypherDriver, uuid string) error {

	query := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			DELETE p
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	return annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
}
