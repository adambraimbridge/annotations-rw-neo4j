package annotations

import (
	"fmt"
)

//Annotations represents a collection of Annotation instances
type Annotations []Annotation

//Annotation is the main struct used to create and return structures
type Annotation struct {
	Predicate         string `json:"predicate,omitempty"`
	ID                string `json:"id,omitempty"`
	AnnotatedBy       string `json:"annotatedBy,omitempty"`
	AnnotatedDate     string `json:"annotatedDate,omitempty"`
	OriginatingSystem string `json:"originatingSystem,omitempty"`
}

const (
	mentionsPred     = "http://www.ft.com/ontology/annotation/mentions"
	aboutPred        = "http://www.ft.com/ontology/annotation/about"
	annotatedByPred  = "http://www.ft.com/ontology/annotation/isAnnotatedBy"
	describesPred    = "http://www.ft.com/ontology/annotation/describes"
	classifiedByPred = "http://www.ft.com/ontology/annotation/isClassifiedBy"
	mentionsRel      = "MENTIONS"
	aboutRel         = "ABOUT"
	describesRel     = "DESCRIBES"
	classifiedByRel  = "CLASSIFIED_BY"
	annotatedByRel   = "ANNOTATED_BY"
)

var neoAnnotationRelationships = []string{
	mentionsRel, aboutRel, describesRel, classifiedByRel, annotatedByRel,
}

var neoTypesToPredicate = map[string]string{
	mentionsRel:     mentionsPred,
	aboutRel:        aboutPred,
	annotatedByRel:  annotatedByPred,
	describesRel:    describesPred,
	classifiedByRel: classifiedByPred,
}

var predicatesToNeoRelationship = map[string]string{
	mentionsPred:     mentionsRel,
	aboutPred:        aboutRel,
	annotatedByPred:  annotatedByRel,
	describesPred:    describesRel,
	classifiedByPred: classifiedByRel,
}

var relationshipInheritance = map[string]string{
	mentionsPred:  annotatedByPred,
	aboutPred:     annotatedByPred,
	describesPred: annotatedByPred,
}

func predicateToNeoType(predicate string) (neoType string) {
	validatePredicate(predicate)
	return predicatesToNeoRelationship[predicate]
}

func validatePredicate(predicate string) (err error) {
	_, exists := predicatesToNeoRelationship[predicate]
	if exists {
		return nil
	}
	return fmt.Errorf("Annotation type URI %s is not one of the regognised ones %+v", predicate, predicatesToNeoRelationship)
}
