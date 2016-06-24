package annotations

import (
	"fmt"
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var annotationsDriver service

const (
	contentUUID       = "32b089d2-2aae-403d-be6e-877404f586cf"
	conceptUUID       = "412e4ca3-f8d5-4456-8606-064c1dba3c45"
	secondConceptUUID = "c834adfa-10c9-4748-8a21-c08537172706"
	oldConceptUUID    = "ad28ddc7-4743-4ed3-9fad-5012b61fb919"
	brandUUID         = "8e21cbd4-e94b-497a-a43b-5b2309badeb3"
	v2PlatformVersion = "v2"
	v1PlatformVersion = "v1"
)

func getURI(uuid string) string {
	return fmt.Sprintf("http://api.ft.com/things/%s", uuid)
}

func TestDeleteRemovesAnnotationsButNotConceptsOrContent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	annotationsToDelete := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToDelete), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToDelete)

	found, err := annotationsDriver.Delete(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotation for content uuid %, conceptUUID %s", contentUUID, conceptUUID)

	anns, found, err := annotationsDriver.Read(contentUUID)

	assert.Equal(annotations{}, anns, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.False(found, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.NoError(err, "Error trying to find annotation for content %s", contentUUID)

	checkNodeIsStillPresent(contentUUID, t)
	checkNodeIsStillPresent(conceptUUID, t)

	err = deleteNode(annotationsDriver, contentUUID)
	assert.NoError(err, "Error trying to delete content node with uuid %s, err=%v", contentUUID, err)
	err = deleteNode(annotationsDriver, conceptUUID)
	assert.NoError(err, "Error trying to delete concept node with uuid %s, err=%v", conceptUUID, err)

}

func TestWriteFailsWhenNoConceptIDSupplied(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	annotationsToWrite := annotations{annotation{
		Thing: thing{PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	err := annotationsDriver.Write(contentUUID, annotationsToWrite)
	assert.Error(err, "Should have failed to write annotation")
	_, ok := err.(ValidationError)
	assert.True(ok, "Should have returned a validation error")
}

func TestWriteAllValuesPresent(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	annotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
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

func TestWriteDoesNotRemoveExistingIsClassifedByBrandRelationships(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	createBrandQuery := &neoism.CypherQuery{
		Statement: `MERGE (b:Brand{uuid:{brandUuid}}) SET b :Concept:Thing RETURN b.uuid`,
		Parameters: map[string]interface{}{
			"brandUuid": brandUUID,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{createBrandQuery})

	createContentQuery := &neoism.CypherQuery{
		Statement: `MERGE (c:Content{uuid:{contentUuid}}) SET c :Thing RETURN c.uuid`,
		Parameters: map[string]interface{}{
			"contentUuid": contentUUID,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{createContentQuery})

	contentQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid:{contentUuid}})
		MERGE (b:Brand{uuid:{brandUuid}})
		CREATE (n)-[rel:IS_CLASSIFIED_BY{platformVersion:{platformVersion}}]->(b) RETURN rel.platformVersion`,
		Parameters: map[string]interface{}{
			"contentUuid":     contentUUID,
			"brandUuid":       brandUUID,
			"platformVersion": v2PlatformVersion,
		},
	}

	err := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{contentQuery})

	annotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")
	found, err := annotationsDriver.Delete(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotations for content uuid %s", contentUUID)

	result := []struct {
		Uuid string `json:"b.uuid"`
	}{}

	getContentQuery := &neoism.CypherQuery{
		Statement: `MATCH (n:Thing {uuid:{contentUuid}})-[:IS_CLASSIFIED_BY]->(b:Brand) RETURN b.uuid`,
		Parameters: map[string]interface{}{
			"contentUuid": contentUUID,
			"brandUuid":   brandUUID,
		},
		Result: &result,
	}

	readErr := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{getContentQuery})
	assert.NoError(readErr)
	assert.NotEmpty(result)

	removeRelationshipQuery := &neoism.CypherQuery{
		Statement: `
			MATCH (b:Thing {uuid:{brandUuid}})<-[rel:IS_CLASSIFIED_BY]-(t:Thing)
			DELETE rel
		`,
		Parameters: map[string]interface{}{
			"brandUuid": brandUUID,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{removeRelationshipQuery})

	err = deleteNode(annotationsDriver, brandUUID)
	assert.NoError(err, "Error trying to delete concept node with uuid %s, err=%v", brandUUID, err)
}

func TestWriteDoesRemoveExistingIsClassifedForV1TermsAndTheirRelationships(t *testing.T) {
	assert := assert.New(t)

	v1AnnotationsDriver := getAnnotationsService(t, v1PlatformVersion)

	createContentQuery := &neoism.CypherQuery{
		Statement: `MERGE (c:Content{uuid:{contentUuid}}) SET c :Thing RETURN c.uuid`,
		Parameters: map[string]interface{}{
			"contentUuid": contentUUID,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{createContentQuery})

	contentQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid:{contentUuid}})
		 	    MERGE (a:Thing{uuid:{conceptUUID}})
			    CREATE (n)-[rel1:MENTIONS{platformVersion:"v2"}]->(a)
			    MERGE (b:Thing{uuid:{secondConceptUUID}})
			    CREATE (n)-[rel2:IS_CLASSIFIED_BY{platformVersion:{platformVersion}}]->(b)`,
		Parameters: map[string]interface{}{
			"contentUuid":       contentUUID,
			"conceptUUID":       conceptUUID,
			"secondConceptUUID": secondConceptUUID,
			"platformVersion":   v1PlatformVersion,
		},
	}

	err := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{contentQuery})

	annotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(v1AnnotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")
	found, err := v1AnnotationsDriver.Delete(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotations for content uuid %s", contentUUID)

	result := []struct {
		Uuid string `json:"b.uuid"`
	}{}

	//CHECK THAT ALL THE v1 annotations were updated
	getContentQuery := &neoism.CypherQuery{
		Statement: `MATCH (n:Thing {uuid:{contentUuid}})-[r]->(b:Thing) where r.platformVersion={platformVersion} RETURN b.uuid`,
		Parameters: map[string]interface{}{
			"contentUuid":     contentUUID,
			"platformVersion": v1PlatformVersion,
		},
		Result: &result,
	}

	readErr := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{getContentQuery})
	assert.NoError(readErr)
	assert.Empty(result)

	//CHECK THAT V2 annotations were not deleted
	getContentQuery = &neoism.CypherQuery{
		Statement: `MATCH (n:Thing {uuid:{contentUuid}})-[r]->(b:Thing) where r.platformVersion={platformVersion} RETURN b.uuid`,
		Parameters: map[string]interface{}{
			"contentUuid":     contentUUID,
			"platformVersion": v2PlatformVersion,
		},
		Result: &result,
	}

	readErr = annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{getContentQuery})
	assert.NoError(readErr)
	assert.NotEmpty(result)

	//Delete v2 annotations
	removeRelationshipQuery := &neoism.CypherQuery{
		Statement: `
			MATCH (b:Thing {uuid:{conceptUUID}})<-[rel]-(t:Thing)
			where rel.platformVersion = "v2"
			DELETE rel
		`,
		Parameters: map[string]interface{}{
			"conceptUUID": conceptUUID,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{removeRelationshipQuery})

	err = deleteNode(annotationsDriver, brandUUID)
	assert.NoError(err, "Error trying to delete concept node with uuid %s, err=%v", brandUUID, err)
	err = deleteNode(annotationsDriver, secondConceptUUID)
	assert.NoError(err, "Error trying to delete concept node with uuid %s, err=%v", secondConceptUUID, err)
}

func TestWriteAndReadMultipleAnnotations(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	annotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}, annotation{
		Thing: thing{ID: getURI(secondConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.4},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.5},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToWrite)

	cleanUp(t, contentUUID, []string{conceptUUID, secondConceptUUID})
}

func TestIfProvenanceGetsWrittenWithEmptyAgentRoleAndTimeValues(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	annotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "",
				AtTime:    "",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, annotationsToWrite), "Failed to write annotation")

	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, annotationsToWrite)

	cleanUp(t, contentUUID, []string{conceptUUID})
}

func TestUpdateWillRemovePreviousAnnotations(t *testing.T) {
	assert := assert.New(t)

	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)

	oldAnnotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, oldAnnotationsToWrite), "Failed to write annotations")
	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, oldAnnotationsToWrite)

	updatedAnnotationsToWrite := annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}

	assert.NoError(annotationsDriver.Write(contentUUID, updatedAnnotationsToWrite), "Failed to write updated annotations")
	readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t, contentUUID, updatedAnnotationsToWrite)

	cleanUp(t, contentUUID, []string{conceptUUID, oldConceptUUID})
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
	err := annotationsDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func TestCreateAnnotationQuery(t *testing.T) {
	assert := assert.New(t)
	annotationToWrite := annotation{
		Thing: thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2PlatformVersion)
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
	}

	for _, test := range tests {
		actualRelationship := getRelationshipFromPredicate(test.predicate)
		if test.relationship != actualRelationship {
			t.Errorf("\nExpected: %s\nActual: %s", test.relationship, actualRelationship)
		}
	}
}

func TestCreateAnnotationQueryWithPredicate(t *testing.T) {
	assert := assert.New(t)
	annotationToWrite := annotation{
		Thing: thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "isClassifiedBy",
		},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}

	query, err := createAnnotationQuery(contentUUID, annotationToWrite, v2PlatformVersion)
	assert.NoError(err, "Cypher query for creating annotations couldn't be created.")
	assert.Contains(query.Statement, "IS_CLASSIFIED_BY", fmt.Sprintf("\nRelationship name is not inserted!"))
	assert.NotContains(query.Statement, "MENTIONS", fmt.Sprintf("\nDefault relationship was insterted insted of IS_CLASSIFIED_BY!"))
}

func getAnnotationsService(t *testing.T, platformVersion string) service {
	assert := assert.New(t)
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return NewAnnotationsService(neoutils.StringerDb{db}, db, platformVersion)
}

func readAnnotationsForContentUUIDAndCheckKeyFieldsMatch(t *testing.T, contentUUID string, expectedAnnotations []annotation) {
	assert := assert.New(t)
	storedThings, found, err := annotationsDriver.Read(contentUUID)
	storedAnnotations := storedThings.([]annotation)

	assert.NoError(err, "Error finding annotations for contentUUID %s", contentUUID)
	assert.True(found, "Didn't find annotations for contentUUID %s", contentUUID)
	assert.Equal(len(expectedAnnotations), len(storedAnnotations), "Didn't get the same number of annotations")
	for idx, expectedAnnotation := range expectedAnnotations {
		storedAnnotation := storedAnnotations[idx]
		assert.EqualValues(expectedAnnotation.Provenances, storedAnnotation.Provenances, "Provenances not the same")

		// In annotations write, we don't store anything other than ID for the concept (so type will only be 'Thing' and pref label will not
		// be present UNLESS the concept has been written by some other system)
		assert.Equal(expectedAnnotation.Thing.ID, storedAnnotation.Thing.ID, "Thing ID not the same")
	}
}

func checkNodeIsStillPresent(uuid string, t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
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
	found, err := annotationsDriver.Delete(contentUUID)
	assert.True(found, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotations for content uuid %s", contentUUID)

	err = deleteNode(annotationsDriver, contentUUID)
	assert.NoError(err, "Could not delete content node")

	for _, conceptUUID := range conceptUUIDs {
		err = deleteNode(annotationsDriver, conceptUUID)
		assert.NoError(err, "Could not delete concept node")
	}
}

func deleteNode(annotationsDriver service, uuid string) error {

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
