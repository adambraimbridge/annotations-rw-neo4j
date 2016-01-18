package annotations

import (
	"fmt"
)

//Annotations represents a collection of Annotation instances
type Annotations []Annotation

//Annotation is the main struct used to return results to the consumers of this api
//Based on com.ft.annotations.mode.AnnotationReadResult in git.svc.ft.com/cp/annotations-api @ 29cab5224bcf5c219e23c5aa2f0446e6de5b4be4
type Annotation struct {
	ID          string       `json:"id"`
	APIURL      string       `json:"apiUrl"`
	Identifiers []Identifier `json:"identifiers"`
	Label       string       `json:"label"`
	Types       []string     `json:"types"`
	LEICode     string       `json:"leiCode,ommitempty"`
}

//Identifier represents the authority that provided the metadata used by an instance of an annotation
type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

var neoTypesToTypeURI = map[string]string{
	"Mentions":      "http://www.ft.com/ontology/annotation/mentions",
	"About":         "http://www.ft.com/ontology/annotation/about",
	"IsAnnotatedBy": "http://www.ft.com/ontology/annotation/isAnnotatedBy",
	"Describes":     "http://www.ft.com/ontology/annotation/describes",
}

var typeURIToNeoType = map[string]string{
	"http://www.ft.com/ontology/annotation/mentions":      "Mentions",
	"http://www.ft.com/ontology/annotation/about":         "About",
	"http://www.ft.com/ontology/annotation/isAnnotatedBy": "IsAnnotatedBy",
	"http://www.ft.com/ontology/annotation/describes":     "Describes",
}

func validateTypeURI(typeToValidate string) (err error) {
	_, exists := typeURIToNeoType[typeToValidate]
	if exists {
		return nil
	}
	return fmt.Errorf("Annotation type URI %s is not one of the regognised ones %+v", typeToValidate, typeURIToNeoType)
}

const (
	fsAuthority = "http://api.ft.com/system/FACTSET-PPL"
)
