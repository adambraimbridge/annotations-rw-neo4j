package annotations

import (
	"errors"
	"fmt"
	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

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

func annotationRelationship(predicate string, idx int) (statement string) {
	stmt := `MERGE (content:Thing{uuid:contentID})-(%s:%s)->(concept:Thing{uuid:%d})`
	statement = fmt.Sprintf(stmt, `aRel`+string(idx), predicateToNeoType(predicate), `conceptID`+string(idx))
	statement = statement + "\n" + fmt.Sprintf(stmt, `aRel`+string(idx), `ANNOTATED_BY`, `conceptID`+string(idx))
	return statement
}

//Create a set of annotations associated with a piece of content
func (driver CypherDriver) Create(contentUUID string, annotations Annotations) (err error) {
	var statement string
	params := neoism.Props{"contentID": contentUUID}
	for idx, annotation := range annotations {
		statement += annotationRelationship(annotation.Predicate, idx)
		params[`conceptID`+string(idx)] = annotation.ID
	}
	query := &neoism.CypherQuery{
		Statement:  statement,
		Parameters: params,
	}
	log.Debugf("Create Annotation for content uuid: %s query: %+v", contentUUID, query)
	return driver.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
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
