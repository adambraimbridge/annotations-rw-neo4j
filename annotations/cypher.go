package annotations

import (
	"errors"
	"fmt"
	"time"

	"regexp"

	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

var uuidExtractRegex = regexp.MustCompile(".*/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$")

// Driver interface
type Driver interface {
	Read(contentUUID string) (annotations Annotations, found bool, err error)
	DeleteAll(contentUUID string) (found bool, err error)
	Write(contentUUID string, annotations Annotations) (err error)
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

func (driver CypherDriver) Read(contentUUID string) (annotations Annotations, found bool, err error) {
	results := []struct {
		Annotations
	}{}

	//TODO shouldn't return Provenances if none of the scores, agentRole or atTime are set
	statementTemplate := `
					MATCH (c:Thing{uuid:{contentUUID}})-[m:MENTIONS]->(cc:Thing)
					WITH c, cc, m, {id:cc.uuid,prefLabel:cc.prefLabel,types:labels(cc)} as t,
					collect(
						{scores:[
							{scoringSystem:'%s', value:m.relevanceScore},
							{scoringSystem:'%s', value:m.confidenceScore}],
						agentRole:m.annotatedBy,
						atTime:m.annotatedDate}) as p
					RETURN [{thing:t, provenances: p}] as annotations
									`
	statement := fmt.Sprintf(statementTemplate, relevanceScoringSystem, confidenceScoringSystem)

	query := &neoism.CypherQuery{
		Statement:  statement,
		Parameters: neoism.Props{"contentUUID": contentUUID},
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
	}
	annotations = results[0].Annotations
	for idx := range annotations {
		mapToResponseFormat(&annotations[idx])
	}

	log.Debugf("Returning %v", annotations)
	return annotations, true, nil
}

//DeleteAll removes all the annotations for this content. Ignore the nodes on either end -
//may leave nodes that are only 'things' inserted by this writer: clean up
//as a result of this will need to happen externally if required
func (driver CypherDriver) DeleteAll(contentUUID string) (bool, error) {

	query := &neoism.CypherQuery{
		Statement:    `MATCH (c:Thing{uuid: {contentUUID}})-[m:MENTIONS]->(cc:Thing) DELETE m`,
		Parameters:   neoism.Props{"contentUUID": contentUUID},
		IncludeStats: true,
	}

	err := driver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	stats, err := query.Stats()
	if err != nil {
		return false, err
	}

	var found bool
	if stats.ContainsUpdates {
		found = true
	}

	return found, err
}

//Write a set of annotations associated with a piece of content. Any annotations
//already there will be removed
func (driver CypherDriver) Write(contentUUID string, annotations Annotations) (err error) {
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

func createAnnotationQuery(contentUUID string, annotation Annotation) (*neoism.CypherQuery, error) {
	query := neoism.CypherQuery{}
	thingID, err := extractUUIDFromURI(annotation.Thing.ID)
	if err != nil {
		return nil, err
	}

	if len(annotation.Provenances) > 1 {
		return nil, errors.New("Cannot insert a MENTIONS annotation with multiple provenances")
	}

	var provenance Provenance
	params := map[string]interface{}{}
	if len(annotation.Provenances) == 1 {
		provenance = annotation.Provenances[0]
		annotatedBy, annotatedDateEpoch, relevanceScore, confidenceScore, supplied, err := extractDataFromProvenance(&provenance)

		if err != nil {
			return nil, err
		}

		if supplied == true {
			params["annotatedBy"] = annotatedBy
			params["annotatedDateEpoch"] = annotatedDateEpoch
			params["relevanceScore"] = relevanceScore
			params["confidenceScore"] = confidenceScore
			params["annotatedDate"] = provenance.AtTime
		}

	}

	query.Statement = createAnnotationRelationship()
	query.Parameters = map[string]interface{}{
		"contentID": contentUUID,
		"conceptID": thingID,
		"annProps":  params,
	}
	return &query, nil
}

func extractDataFromProvenance(provenance *Provenance) (string, int64, float64, float64, bool, error) {
	if provenance.AgentRole == "" || provenance.AtTime == "" || len(provenance.Scores) == 0 {
		return "", -1, -1, -1, false, nil
	}
	var annotatedBy string
	var annotatedDateEpoch int64
	var confidenceScore, relevanceScore float64
	var err error
	annotatedBy, err = extractUUIDFromURI(provenance.AgentRole)
	annotatedDateEpoch, err = convertAnnotatedDateToEpoch(provenance.AtTime)
	relevanceScore, confidenceScore, err = extractScores(provenance.Scores)

	if err != nil {
		return "", -1, -1, -1, true, err
	}
	return annotatedBy, annotatedDateEpoch, relevanceScore, confidenceScore, true, nil
}

func extractUUIDFromURI(uri string) (string, error) {
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

func validateAnnotations(annotations *Annotations) error {
	//TODO - for consistency, we should probably just not create the annotation?
	for _, annotation := range *annotations {
		if annotation.Thing.ID == "" {
			return fmt.Errorf("Concept uuid missing for annotation %+v", annotation)
		}
	}
	return nil
}

func mapToResponseFormat(annotation *Annotation) {
	annotation.Thing.ID = mapper.IDURL(annotation.Thing.ID)
	// We expect only ONE provenance
	var provenanceValid bool
	for idx := range annotation.Provenances {
		if annotation.Provenances[idx].AgentRole != "" {
			annotation.Provenances[idx].AgentRole = mapper.IDURL(annotation.Provenances[idx].AgentRole)
			provenanceValid = true
		}
	}
	if provenanceValid != true {
		annotation.Provenances = nil
	}
}
