package annotations

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"regexp"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

var uuidExtractRegex = regexp.MustCompile(".*/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$")

// Service interface. Compatible with the baserwftapp service EXCEPT for
// 1) the Write function, which has signature Write(thing interface{}) error...
// 2) the DecodeJson function, which has signature DecodeJSON(*json.Decoder) (thing interface{}, identity string, err error)
// The problem is that we have a list of things, and the uuid is for a related OTHER thing
// TODO - move to implement a shared defined Service interface?
type Service interface {
	Write(contentUUID string, thing interface{}) (err error)
	Read(contentUUID string) (thing interface{}, found bool, err error)
	Delete(contentUUID string) (found bool, err error)
	Check() (err error)
	DecodeJSON(*json.Decoder) (thing interface{}, err error)
	Count() (int, error)
	Initialise() error
}

//holds the Neo4j-specific information
type service struct {
	cypherRunner    neoutils.CypherRunner
	indexManager    neoutils.IndexManager
	platformVersion string
}

//NewAnnotationsService instantiate driver
func NewAnnotationsService(cypherRunner neoutils.CypherRunner, indexManager neoutils.IndexManager, platformVersion string) service {
	if platformVersion == "" {
		log.Fatalf("PlatformVersion was not specified!")
	}
	return service{cypherRunner, indexManager, platformVersion}
}

// DecodeJSON decodes to a list of annotations, for ease of use this is a struct itself
func (s service) DecodeJSON(dec *json.Decoder) (interface{}, error) {
	a := annotations{}
	err := dec.Decode(&a)
	return a, err
}

func (s service) Read(contentUUID string) (thing interface{}, found bool, err error) {
	results := []annotation{}

	//TODO shouldn't return Provenances if none of the scores, agentRole or atTime are set
	statementTemplate := `
					MATCH (c:Thing{uuid:{contentUUID}})-[rel{platformVersion:{platformVersion}}]->(cc:Thing)
					WHERE rel.lifecycle = {lifecycle}
					OR rel.lifecycle IS NULL
					WITH c, cc, rel, {id:cc.uuid,prefLabel:cc.prefLabel,types:labels(cc),predicate:type(rel)} as thing,
					collect(
						{scores:[
							{scoringSystem:'%s', value:rel.relevanceScore},
							{scoringSystem:'%s', value:rel.confidenceScore}],
						agentRole:rel.annotatedBy,
						atTime:rel.annotatedDate}) as provenances
					RETURN thing, provenances ORDER BY thing.id
									`
	statement := fmt.Sprintf(statementTemplate, relevanceScoringSystem, confidenceScoringSystem)

	query := &neoism.CypherQuery{
		Statement:  statement,
		Parameters: neoism.Props{"contentUUID": contentUUID, "platformVersion": s.platformVersion, "lifecycle": "annotations-" + s.platformVersion},
		Result:     &results,
	}
	err = s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("CypherResult Read Annotations for uuid: %s was: %+v", contentUUID, results)
	if (len(results)) == 0 {
		return annotations{}, false, nil
	}

	for idx := range results {
		mapToResponseFormat(&results[idx])
	}

	return results, true, nil
}

//Delete removes all the annotations for this content. Ignore the nodes on either end -
//may leave nodes that are only 'things' inserted by this writer: clean up
//as a result of this will need to happen externally if required
func (s service) Delete(contentUUID string) (bool, error) {
	deleteStatement := `MATCH (c:Thing{uuid: {contentUUID}})-[rel{platformVersion:{platformVersion}}]->(cc:Thing)
											WHERE rel.lifecycle = {lifecycle}
											OR rel.lifecycle IS NULL
	 										DELETE rel`
	query := &neoism.CypherQuery{
		Statement:    deleteStatement,
		Parameters:   neoism.Props{"contentUUID": contentUUID, "platformVersion": s.platformVersion, "lifecycle": "annotations-" + s.platformVersion},
		IncludeStats: true,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

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
func (s service) Write(contentUUID string, thing interface{}) (err error) {
	annotationsToWrite := thing.(annotations)

	if contentUUID == "" {
		return errors.New("Content uuid is required")
	}
	if err := validateAnnotations(&annotationsToWrite); err != nil {
		log.Warnf("Validation of supplied annotations failed")
		return err
	}

	if len(annotationsToWrite) == 0 {
		log.Warnf("No new annotations supplied for content uuid: %s", contentUUID)
	}

	queries := append([]*neoism.CypherQuery{}, dropAllAnnotationsQuery(contentUUID, s.platformVersion))

	var statements = []string{}
	for _, annotationToWrite := range annotationsToWrite {
		query, err := createAnnotationQuery(contentUUID, annotationToWrite, s.platformVersion)
		if err != nil {
			return err
		}
		statements = append(statements, query.Statement)
		queries = append(queries, query)
	}
	log.Infof("Updated Annotations for content uuid: %s", contentUUID)
	log.Debugf("For update, ran statements: %+v", statements)

	return s.cypherRunner.CypherBatch(queries)
}

// Check tests neo4j by running a simple cypher query
func (s service) Check() error {
	return neoutils.Check(s.cypherRunner)
}

func (s service) Count() (int, error) {
	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH ()-[r{platformVersion:{platformVersion}}]->()
								WHERE rel.lifecycle = {lifecycle}
								OR rel.lifecycle IS NULL
								RETURN count(r) as c`,
		Parameters: neoism.Props{"platformVersion": s.platformVersion, "lifecycle": "annotations-" + s.platformVersion},
		Result:     &results,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}

func (s service) Initialise() error {
	return nil // No constraints need to be set up
}

func createAnnotationRelationship(relation string) (statement string) {
	stmt := `
                MERGE (content:Thing{uuid:{contentID}})
                MERGE (upp:Identifier:UPPIdentifier{value:{conceptID}})
                MERGE (upp)-[:IDENTIFIES]->(concept:Thing) ON CREATE SET concept.uuid = {conceptID}
                MERGE (content)-[pred:%s {platformVersion:{platformVersion}, lifecycle: {lifecycle}}]->(concept)
                SET pred={annProps}
          `
	statement = fmt.Sprintf(stmt, relation)
	return statement
}

func getRelationshipFromPredicate(predicate string) (relation string) {
	if predicate != "" {
		relation = relations[predicate]
	} else {
		relation = relations["mentions"]
	}
	return relation
}

func createAnnotationQuery(contentUUID string, ann annotation, platformVersion string) (*neoism.CypherQuery, error) {
	query := neoism.CypherQuery{}
	thingID, err := extractUUIDFromURI(ann.Thing.ID)
	if err != nil {
		return nil, err
	}

	//todo temporary change to deal with multiple provenances
	/*if len(ann.Provenances) > 1 {
		return nil, errors.New("Cannot insert a MENTIONS annotation with multiple provenances")
	}*/

	var prov provenance
	params := map[string]interface{}{}
	params["platformVersion"] = platformVersion

	if len(ann.Provenances) >= 1 {
		prov = ann.Provenances[0]
		annotatedBy, annotatedDateEpoch, relevanceScore, confidenceScore, supplied, err := extractDataFromProvenance(&prov)

		if err != nil {
			log.Infof("ERROR=%s", err)
			return nil, err
		}

		if supplied == true {
			if annotatedBy != "" {
				params["annotatedBy"] = annotatedBy
			}
			if prov.AtTime != "" {
				params["annotatedDateEpoch"] = annotatedDateEpoch
				params["annotatedDate"] = prov.AtTime
			}
			params["relevanceScore"] = relevanceScore
			params["confidenceScore"] = confidenceScore
		}
	}

	relation := getRelationshipFromPredicate(ann.Thing.Predicate)
	query.Statement = createAnnotationRelationship(relation)
	query.Parameters = map[string]interface{}{
		"contentID":       contentUUID,
		"conceptID":       thingID,
		"platformVersion": platformVersion,
		"lifecycle":       "annotations-" + platformVersion,
		"annProps":        params,
	}
	return &query, nil
}

func extractDataFromProvenance(prov *provenance) (string, int64, float64, float64, bool, error) {
	if len(prov.Scores) == 0 {
		return "", -1, -1, -1, false, nil
	}
	var annotatedBy string
	var annotatedDateEpoch int64
	var confidenceScore, relevanceScore float64
	var err error
	if prov.AgentRole != "" {
		annotatedBy, err = extractUUIDFromURI(prov.AgentRole)
	}
	if prov.AtTime != "" {
		annotatedDateEpoch, err = convertAnnotatedDateToEpoch(prov.AtTime)
	}
	relevanceScore, confidenceScore, err = extractScores(prov.Scores)

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

func extractScores(scores []score) (float64, float64, error) {
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

func dropAllAnnotationsQuery(contentUUID string, platformVersion string) *neoism.CypherQuery {
	matchStmtTemplate := `OPTIONAL MATCH (:Thing{uuid:{contentID}})-[r {platformVersion:{platformVersion}, lifecycle: {lifecycle}}]->(t:Thing)
											DELETE r`

	query := neoism.CypherQuery{}
	query.Statement = matchStmtTemplate
	query.Parameters = neoism.Props{"contentID": contentUUID, "platformVersion": platformVersion, "lifecycle": "annotations-" + platformVersion}
	return &query
}

func validateAnnotations(annotations *annotations) error {
	//TODO - for consistency, we should probably just not create the annotation?
	for _, annotation := range *annotations {
		if annotation.Thing.ID == "" {
			return ValidationError{fmt.Sprintf("Concept uuid missing for annotation %+v", annotation)}
		}
	}
	return nil
}

//ValidationError is thrown when the annotations are not valid because mandatory information is missing
type ValidationError struct {
	Msg string
}

func (v ValidationError) Error() string {
	return v.Msg
}

func mapToResponseFormat(ann *annotation) {
	ann.Thing.ID = mapper.IDURL(ann.Thing.ID)
	// We expect only ONE provenance - provenance value is considered valid even if the AgentRole is not specified. See: v1 - isClassifiedBy
	for idx := range ann.Provenances {
		if ann.Provenances[idx].AgentRole != "" {
			ann.Provenances[idx].AgentRole = mapper.IDURL(ann.Provenances[idx].AgentRole)
		}
	}
}
