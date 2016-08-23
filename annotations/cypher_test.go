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
	conceptUUID       = "a7732a22-3884-4bfe-9761-fef161e41d69"
	secondConceptUUID = "c834adfa-10c9-4748-8a21-c08537172706"
	oldConceptUUID    = "ad28ddc7-4743-4ed3-9fad-5012b61fb919"
	brandUUID         = "8e21cbd4-e94b-497a-a43b-5b2309badeb3"
	v2PlatformVersion = "v2"
	v1PlatformVersion = "v1"
	contentLifecyle   = "content"
	annotationsV2     = "annotations-v2"
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

	deleted, err := annotationsDriver.Delete(contentUUID)
	assert.True(deleted, "Didn't manage to delete annotations for content uuid %s", contentUUID)
	assert.NoError(err, "Error deleting annotation for content uuid %, conceptUUID %s", contentUUID, conceptUUID)

	anns, found, err := annotationsDriver.Read(contentUUID)

	assert.Equal(annotations{}, anns, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.False(found, "Found annotation for content %s when it should have been deleted", contentUUID)
	assert.NoError(err, "Error trying to find annotation for content %s", contentUUID)

	checkContentNodeIsStillPresent(contentUUID, t)
	checkConceptNodeIsStillPresent(conceptUUID, t)

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

func TestWriteDoesNotRemoveExistingIsClassifiedByBrandRelationshipsWithoutLifeCycle(t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
	defer cleanDB(t, assert)

	contentQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid:{contentUuid}}) SET n :Thing
		MERGE (b:Brand{uuid:{brandUuid}}) SET b :Concept:Thing
		CREATE (n)-[rel:IS_CLASSIFIED_BY{platformVersion:{platformVersion}}]->(b)`,
		Parameters: map[string]interface{}{
			"contentUuid":     contentUUID,
			"brandUuid":       brandUUID,
			"platformVersion": v2PlatformVersion,
		},
	}

	annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{contentQuery})

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
	checkRelationship(assert, contentUUID, "v2")

	deleted, err := annotationsDriver.Delete(contentUUID)
	assert.True(deleted, "Didn't manage to delete annotations for content uuid %s", contentUUID)
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
}

func TestWriteDoesNotRemoveExistingIsClassifiedByBrandRelationshipsWithContentLifeCycle(t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
	defer cleanDB(t, assert)
	contentQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid:{contentUuid}}) SET n :Thing
		MERGE (b:Brand{uuid:{brandUuid}}) SET b :Concept:Thing
		CREATE (n)-[rel:IS_CLASSIFIED_BY{platformVersion:{platformVersion}, lifecycle: {lifecycle}}]->(b)`,
		Parameters: map[string]interface{}{
			"contentUuid":     contentUUID,
			"brandUuid":       brandUUID,
			"platformVersion": v2PlatformVersion,
			"lifecycle":       contentLifecyle,
		},
	}

	err := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{contentQuery})
	assert.NoError(err, "Error c for content uuid %s", contentUUID)

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
	deleted, err := annotationsDriver.Delete(contentUUID)
	assert.True(deleted, "Didn't manage to delete annotations for content uuid %s", contentUUID)
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
}

func TestWriteDoesRemoveExistingIsClassifedForV1TermsAndTheirRelationships(t *testing.T) {
	assert := assert.New(t)

	v1AnnotationsDriver := getAnnotationsService(t, v1PlatformVersion)
	annotationsDriver := getAnnotationsService(t, v2PlatformVersion)

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
	assert.NoError(err, "Error writing annotations for content uuid %s", contentUUID)

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
	deleted, err := v1AnnotationsDriver.Delete(contentUUID)
	assert.True(deleted, "Didn't manage to delete annotations for content uuid %s", contentUUID)
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
	storedAnnotations := storedThings.(annotations)

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

func checkContentNodeIsStillPresent(uuid string, t *testing.T) {
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

func checkConceptNodeIsStillPresent(uuid string, t *testing.T) {
	assert := assert.New(t)
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
	results := []struct {
		UUID string `json:"uuid"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Thing)<-[IDENTIFIER]-(upp:UPPIdentifier{value:{uuid}}) return n.uuid
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

func writeClassifedByRelationship(db *neoism.Database, contentId string, conceptId string, lifecycle string, t *testing.T, assert *assert.Assertions) {

	var annotateQuery string
	var qs []*neoism.CypherQuery

	if lifecycle == "" {
		annotateQuery = `
                MERGE (content:Thing{uuid:{contentId}})
                MERGE (upp:Identifier:UPPIdentifier{value:{conceptId}})
                MERGE (upp)-[:IDENTIFIES]->(concept:Thing) ON CREATE SET concept.uuid = {conceptId}
                MERGE (content)-[pred:IS_CLASSIFIED_BY {platformVersion:'v1'}]->(concept)
          `
		qs = []*neoism.CypherQuery{
			{
				Statement:  annotateQuery,
				Parameters: neoism.Props{"contentId": contentId, "conceptId": conceptId},
			},
		}
	} else {
		annotateQuery = `
                MERGE (content:Thing{uuid:{contentId}})
                MERGE (upp:Identifier:UPPIdentifier{value:{conceptId}})
                MERGE (upp)-[:IDENTIFIES]->(concept:Thing) ON CREATE SET concept.uuid = {conceptId}
                MERGE (content)-[pred:IS_CLASSIFIED_BY {platformVersion:'v1', lifecycle: {lifecycle}}]->(concept)
          `
		qs = []*neoism.CypherQuery{
			{
				Statement:  annotateQuery,
				Parameters: neoism.Props{"contentId": contentId, "conceptId": conceptId, "lifecycle": lifecycle},
			},
		}

	}

	err := db.CypherBatch(qs)
	assert.NoError(err)
}

func checkRelationship(assert *assert.Assertions, contentID string, platformVersion string) {
	countQuery := `Match (t:Thing {uuid: {contentID}})-[r {lifecycle: {lifecycle}}]-(x) return count(r) as c`

	results := []struct {
		Count int `json:"c"`
	}{}

	qs := &neoism.CypherQuery{
		Statement:  countQuery,
		Parameters: neoism.Props{"contentID": contentID, "lifecycle": "annotations-" + platformVersion},
		Result:     &results,
	}

	err := annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{qs})
	assert.NoError(err)
	assert.Equal(1, len(results), "More results found than expected!")
	assert.Equal(1, results[0].Count, "No Relationship with Lifecycle found!")
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

/*func checkDbClean(db *neoism.Database, t *testing.T) {
	assert := assert.New(t)

	result := []struct {
		Uuid string `json:"org.uuid"`
	}{}

	checkGraph := neoism.CypherQuery{
		Statement: `
			MATCH (org:Thing) WHERE org.uuid in {uuids} RETURN org.uuid
		`,
		Parameters: neoism.Props{
			"uuids": []string{fullContentUuid, minimalContentUuid},
		},
		Result: &result,
	}
	err := db.Cypher(&checkGraph)
	assert.NoError(err)
	assert.Empty(result)
}*/

func cleanDB(t *testing.T, assert *assert.Assertions) {
	annotationsDriver = getAnnotationsService(t, v2PlatformVersion)
	qs := []*neoism.CypherQuery{
		{
			Statement: fmt.Sprintf("MATCH (mc:Thing {uuid: '%v'}) DETACH DELETE mc", contentUUID),
		},
		{
			Statement: fmt.Sprintf("MATCH (fc:Thing {uuid: '%v'}) DETACH DELETE fc", conceptUUID),
		},
		{
			Statement: fmt.Sprintf("MATCH (fc:Thing {uuid: '%v'}) DETACH DELETE fc", secondConceptUUID),
		},
		{
			Statement: fmt.Sprintf("MATCH (fc:Thing {uuid: '%v'}) DETACH DELETE fc", oldConceptUUID),
		},
		{
			Statement: fmt.Sprintf("MATCH (fc:Thing {uuid: '%v'}) DETACH DELETE fc", brandUUID),
		},
	}

	err := annotationsDriver.cypherRunner.CypherBatch(qs)
	assert.NoError(err)
}

func deleteNode(annotationsDriver service, uuid string) error {

	query := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			OPTIONAL MATCH (identifier:UPPIdentifier)-[rel:IDENTIFIES]->(p)
			DELETE identifier, rel, p
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	return annotationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
}
