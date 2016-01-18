package annotations

import (
	"encoding/json"
	"errors"
	"fmt"
"strings"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
)

//CypherDriver connectivity to neo4j datastore
type CypherDriver struct {
	cypherRunner neocypherrunner.CypherRunner
	indexManager neoutils.IndexManager
}

//NewCypherDriver constructor for the CypherDriver
func NewCypherDriver(cypherRunner neocypherrunner.CypherRunner, indexManager neoutils.IndexManager) CypherDriver {
	return CypherDriver{cypherRunner, indexManager}
}

//Initialise ensures the required indexes have been created
func (pcd CypherDriver) Initialise() error {
	return neoutils.EnsureIndexes(pcd.indexManager, map[string]string{
		"Thing":   "uuid",
		"Concept": "uuid",
		"Content": "uuid"})
}

func (pcd CypherDriver) Read(uuid string) (interface{}, bool, error) {
	results := []struct {
		UUID              string `json:"uuid"`
		Name              string `json:"name"`
		BirthYear         int    `json:"birthYear"`
		Salutation        string `json:"salutation"`
		FactsetIdentifier string `json:"factsetIdentifier"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Person {uuid:{uuid}}) return n.uuid
		as uuid, n.name as name,
		n.factsetIdentifier as factsetIdentifier,
		n.birthYear as birthYear,
		n.salutation as salutation`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	err := pcd.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return Annotations{}, false, err
	}

	if len(results) == 0 {
		return Annotations{}, false, nil
	}

	result := results[0]

	p := Annotations{}

	if result.FactsetIdentifier != "" {
		//		p.Identifiers = append(p.Identifiers, identifier{fsAuthority, result.FactsetIdentifier})
	}

	return p, true, nil

}

func writeQueryStatements() (statements map[string]string) {
	mergeContent := `MERGE (content:Thing{uuid:contentID})-`
	concept := `->(concept:Thing{uuid:conceptID})`
	statements = map[string]string{
		"Mentions":      mergeContent + `(rel:MENTIONS)` + concept,
		"About":         mergeContent + `(rel:ABOUT)` + concept,
		"IsAnnotatedBy": mergeContent + `(rel:ANNOTATED_BY)` + concept,
		"Describes":     mergeContent + `(rel:DESCRIBES)` + concept,
	}
	return statements
}

func (pcd CypherDriver) Write(thing interface{}) error {
	queryStatements := writeQueryStatements()
	annotations := thing.(Annotations)

	query := &neoism.CypherQuery{
		Parameters: map[string]interface{}{
			"contentID": ,
		},
	}

	for _, annotation := range annotations {
		params := map[string]interface{}{}

	}

	return pcd.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

}

//Delete removes all annotations on a document
func (pcd CypherDriver) Delete(uuid string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			REMOVE p:Concept
			REMOVE p:Person
			SET p={props}
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			OPTIONAL MATCH (p)-[a]-(x)
			WITH p, count(a) AS relCount
			WHERE relCount = 0
			DELETE p
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := pcd.cypherRunner.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

	s1, err := clearNode.Stats()
	if err != nil {
		return false, err
	}

	var deleted bool
	if s1.ContainsUpdates && s1.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

//DecodeJSON decodes json !
func (pcd CypherDriver) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	annotation := Annotation{}
	err := dec.Decode(&annotation)
	return annotation, annotation.ID, err

}

// Check ensures the datastore is available
func (pcd CypherDriver) Check() (check v1a.Check) {
	type hcUUIDResult struct {
		UUID string `json:"uuid"`
	}

	checker := func() (string, error) {
		var result []hcUUIDResult

		query := &neoism.CypherQuery{
			Statement: `MATCH (n:Person)
					return  n.uuid as uuid
					limit 1`,
			Result: &result,
		}

		err := pcd.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

		if err != nil {
			return "", err
		}
		if len(result) == 0 {
			return "", errors.New("No Person found")
		}
		if result[0].UUID == "" {
			return "", errors.New("UUID not set")
		}
		return fmt.Sprintf("Found a person with a valid uuid = %v", result[0].UUID), nil
	}

	return v1a.Check{
		BusinessImpact:   "Cannot read/write people via this writer",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: fmt.Sprintf("Cannot connect to Neo4j instance %s with at least one person loaded in it", pcd.cypherRunner),
		Checker:          checker,
	}
}

//Count is a simple stats endpoint that counts the number of annotations
func (pcd CypherDriver) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Person) return count(n) as c`,
		Result:    &results,
	}

	err := pcd.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}
