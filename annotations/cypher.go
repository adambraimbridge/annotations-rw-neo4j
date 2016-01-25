package annotations

import (
	"errors"
	"fmt"
	"time"

	"regexp"

	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

var uuidExtractRegex = regexp.MustCompile(".*/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$")

// Driver interface
type Driver interface {
	Read(contentUUID string) (annotations Annotations, found bool, err error)
	//	Delete(contentUUID string, conceptUUID string) (err error)
	Create(contentUUID string, annotations Annotations) (err error)
	CheckConnectivity() (err error)
}

// CypherDriver struct
type CypherDriver struct {
	cypherRunner neocypherrunner.CypherRunner
	indexManager neoutils.IndexManager
}

//NewCypherDriver instantiate driver
func NewCypherDriver(cypherRunner neocypherrunner.CypherRunner, indexManager neoutils.IndexManager) CypherDriver {
	return CypherDriver{cypherRunner, indexManager}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (driver CypherDriver) CheckConnectivity() (err error) {
	results := []struct {
		ID int
	}{}
	query := &neoism.CypherQuery{
		Statement: "MATCH (x) RETURN ID(x) LIMIT 1",
		Result:    &results,
	}
	err = driver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	log.Debugf("CheckConnectivity results:%+v  err: %+v", results, err)
	return err
}

func createAnnotationRelationship() (statement string) {
	stmt := `
                MERGE (content:Thing{uuid:{contentID}})
                MERGE (concept:Thing{uuid:{conceptID}})
                MERGE (content)-[pred:%s]->(concept)
                SET pred={annProps}
                `
	statement = fmt.Sprintf(stmt, mentionsRel)
	return statement
}

//TODO should we create the Thing with a prefLabel and type if it doesn't exist? Since we know both.
func createAnnotationQuery(contentUUID string, annotation Annotation) (*neoism.CypherQuery, error) {
	query := neoism.CypherQuery{}
	thingID, err := extractUUIDFromUri(annotation.Thing.ID)
	if err != nil {
		return nil, err
	}
	annotatedBy, err := extractUUIDFromUri(annotation.Provenances[0].AgentRole)
	if err != nil {
		return nil, err
	}
	annotatedDateEpoch, err := convertAnnotatedDateToEpoch(annotation.Provenances[0].AtTime)
	if err != nil {
		return nil, err
	}
	relevanceScore, confidenceScore, err := extractScores(annotation.Provenances[0].Scores)

	query.Statement = createAnnotationRelationship()
	//TODO only set the annProps if they are provided
	//TODO need to use the real ID not the supplied uri (i.e. extract the uuid)
	query.Parameters = neoism.Props{
		"contentID": contentUUID,
		"conceptID": thingID,
		"annProps": neoism.Props{
			"annotatedDate":      annotation.Provenances[0].AtTime,
			"annotatedDateEpoch": annotatedDateEpoch,
			"relevanceScore":     relevanceScore,
			"confidenceScore":    confidenceScore,
			"annotatedBy":        annotatedBy,
		},
	}
	return &query, nil
}

func extractUUIDFromUri(uri string) (string, error) {
	result := uuidExtractRegex.FindStringSubmatch(uri)
	if len(result) == 2 {
		return result[1], nil
	}
	return "", fmt.Errorf("Couldn't extract uuid from uri %s", uri)
}

func convertAnnotatedDateToEpoch(annotatedDateString string) (int64, error) {
	datetimeEpoch, err := time.Parse(time.RFC3339, annotatedDateString)

	if err != nil {
		return 0, err
	}

	return datetimeEpoch.Unix(), nil
}

func extractScores(scores []Score) (float64, float64, error) {
	var relevanceScore, confidenceScore float64
	for _, score := range scores {
		scoringSystem := score.ScoringSystem
		value := score.Value
		switch scoringSystem {
		case relevanceScoringSystem:
			relevanceScore = value
		case confidenceScoringSystem:
			confidenceScore = value
		}
	}
	return relevanceScore, confidenceScore, nil
}

func dropAllAnnotationsQuery(contentUUID string) *neoism.CypherQuery {
	matchStmtTemplate := `optional match (:Thing{uuid:{contentID}})-[r:%s]->(:Thing)
                        delete r`

	query := neoism.CypherQuery{}
	query.Statement = fmt.Sprintf(matchStmtTemplate, mentionsRel)
	query.Parameters = neoism.Props{"contentID": contentUUID}
	return &query
}

//Create a set of annotations associated with a piece of content
func (driver CypherDriver) Create(contentUUID string, annotations Annotations) (err error) {
	if contentUUID == "" {
		return errors.New("Content uuid is required")
	}
	if err := validateAnnotations(&annotations); err != nil {
		return fmt.Errorf("Annotation for content %s is not valid. %s", contentUUID, err.Error())
	}
	queries := append([]*neoism.CypherQuery{}, dropAllAnnotationsQuery(contentUUID))
	for _, annotation := range annotations {
		query, err := createAnnotationQuery(contentUUID, annotation)
		if err != nil {
			return err
		}
		queries = append(queries, query)
	}
	log.Debugf("Create Annotation for content uuid: %s query: %+v\n", contentUUID, queries)
	return driver.cypherRunner.CypherBatch(queries)
}

func validateAnnotations(annotations *Annotations) error {
	for _, annotation := range *annotations {
		if annotation.Thing.ID == "" {
			return fmt.Errorf("Concept uuid missing for annotation %+v", annotation)
		}
	}
	return nil
}

type neoChangeEvent struct {
	StartedAt string
	EndedAt   string
}

type neoReadStruct struct {
	P struct {
		ID        string
		Types     []string
		PrefLabel string
		Labels    []string
	}
	M []struct {
		M struct {
			ID           string
			Types        []string
			PrefLabel    string
			Title        string
			ChangeEvents []neoChangeEvent
		}
		O struct {
			ID        string
			Types     []string
			PrefLabel string
			Labels    []string
		}
		R []struct {
			ID           string
			Types        []string
			PrefLabel    string
			ChangeEvents []neoChangeEvent
		}
	}
}

func (driver CypherDriver) Read(contentUUID string) (annotations Annotations, found bool, err error) {
	results := []struct {
		neoReadStruct
	}{}
	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (content:Thing{uuid{:contentUUID}})-[rel]->(concept:Thing)
                        WITH content, collect(type: labels(rel), concept:concept))
                        RETURN collect(content.uuid annotations)
                `,
		Parameters: neoism.Props{"uuid": contentUUID},
		Result:     &results,
	}
	err = driver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v\n", contentUUID, query.Statement, err)
		return Annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("CypherResult Read Annotations for uuid: %s was: %+v", contentUUID, results)
	if (len(results)) == 0 {
		return Annotations{}, false, nil
	} else if len(results) != 1 {
		errMsg := fmt.Sprintf("Multiple people found with the same uuid:%s !", contentUUID)
		log.Error(errMsg)
		return Annotations{}, true, errors.New(errMsg)
	}
	annotations = neoReadStructToAnnotations(results[0].neoReadStruct)
	log.Debugf("Returning %v", annotations)
	return annotations, true, nil
}

func neoReadStructToAnnotations(neo neoReadStruct) (annotations Annotations) {
	log.Debugf("neoReadStructToPerson neo: %+v result: %+v", neo, annotations)
	return annotations
}
