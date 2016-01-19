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

var neoTypesToPredicate = map[string]string{
	"MENTIONS":      "http://www.ft.com/ontology/annotation/mentions",
	"ABOUT":         "http://www.ft.com/ontology/annotation/about",
	"ANNOTATED_BY":  "http://www.ft.com/ontology/annotation/isAnnotatedBy",
	"DESCRIBES":     "http://www.ft.com/ontology/annotation/describes",
	"CLASSIFIED_BY": "http://www.ft.com/ontology/annotation/isClassifiedBy",
}

var predicatesToNeoType = map[string]string{
	"http://www.ft.com/ontology/annotation/mentions":       "MENTIONS",
	"http://www.ft.com/ontology/annotation/about":          "ABOUT",
	"http://www.ft.com/ontology/annotation/isAnnotatedBy":  "ANNOTATED_BY",
	"http://www.ft.com/ontology/annotation/describes":      "DESCRIBES",
	"http://www.ft.com/ontology/annotation/isClassifiedBy": "CLASSIFIED_BY",
}

var requiredPredicates = []string{"ANNOTATED_BY"}

func predicateToNeoType(predicate string) (neoType string) {
	validatePredicate(predicate)
	return predicatesToNeoType[predicate]
}

func validatePredicate(predicate string) (err error) {
	_, exists := predicatesToNeoType[predicate]
	if exists {
		return nil
	}
	return fmt.Errorf("Annotation type URI %s is not one of the regognised ones %+v", predicate, predicatesToNeoType)
}

const (
	fsAuthority = "http://api.ft.com/system/FACTSET-PPL"
)
