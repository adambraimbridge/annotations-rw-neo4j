package annotations

import (
	"errors"
	"fmt"
	"github.com/Financial-Times/neo-cypher-runner-go"
	"github.com/Financial-Times/neo-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
	"time"
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

func annotationRelationship(predicate string) (statement string) {
	stmt := `
                MERGE (content:Thing{uuid:{contentID}})
                MERGE (concept:Thing{uuid:{conceptID}})
                MERGE (content)-[pred:%s]->(concept)
                SET pred={annProps}
                `
	statement = fmt.Sprintf(stmt, predicateToNeoType(predicate))
	log.Debugln(statement)
	return statement
}

//Create a set of annotations associated with a piece of content
func (driver CypherDriver) Create(contentUUID string, annotations Annotations) (err error) {
	if contentUUID == "" {
		return errors.New("Content uuid is required")
	}
	queries := []*neoism.CypherQuery{}
	for _, annotation := range annotations {
		if err := validateAnnotation(&annotation); err != nil {
			return fmt.Errorf("Annotation for content %s is not valid. %s", contentUUID, err.Error())
		}
		query := neoism.CypherQuery{}
		query.Statement = annotationRelationship(annotation.Predicate)
		query.Parameters = neoism.Props{
			"contentID": contentUUID,
			"conceptID": annotation.ID,
			"annProps": neoism.Props{
				"date":      annotation.AnnotatedDate,
				"annotator": annotation.AnnotatedBy,
				"system":    annotation.OriginatingSystem,
			},
		}
		queries = append(queries, &query)
	}
	log.Debugf("Create Annotation for content uuid: %s query: %+v\n", contentUUID, queries)
	return driver.cypherRunner.CypherBatch(queries)
}

func validateAnnotation(annotation *Annotation) error {
	if annotation.Predicate == "" {
		return fmt.Errorf("Predicate missing for annotation %+v", annotation)
	}
	if err := validatePredicate(annotation.Predicate); err != nil {
		return err
	}
	if annotation.ID == "" {
		return fmt.Errorf("Concept uuid missing for annotation %+v", annotation)
	}
	if annotation.AnnotatedDate == "" {
		annotation.AnnotatedDate = time.Now().Format(time.RFC3339)
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
