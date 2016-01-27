package annotations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// AnnotationsDriver for cypher queries
var AnnotationsDriver Driver

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Annotation API requests",
		Name:             "Check connectivity to Neo4j --neo-url is part of the service_args in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance with at least one person loaded in it",
		Checker:          Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := AnnotationsDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func jsonMessage(msgText string) []byte {
	return []byte(fmt.Sprintf(`{"message":"%s"}`, msgText))
}

func isContentTypeJSON(r *http.Request) error {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "application/json") {
		return errors.New("Http Header 'Content-Type' is not 'application/json', this is a JSON API")
	}
	return nil
}

// PutAnnotations handles the replacement of a set of annotations for a given bit of content
func PutAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := isContentTypeJSON(r); err != nil {
		log.Error(err)
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
		return
	}
	var annotations Annotations
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&annotations)
	if err != nil {
		msg := fmt.Sprintf("Error (%v) parsing annotation request %+v", err, r.Body)
		log.Error(msg)
		//TDO - http.Error overwrites content type to text/html, can't use this approach
		http.Error(w, string(jsonMessage(msg)), http.StatusBadRequest)
		return
	}
	err = AnnotationsDriver.Write(uuid, annotations)
	if err != nil {
		msg := fmt.Sprintf("Error creating annotation (%v)", err)
		log.Error(msg)
		http.Error(w, string(jsonMessage(msg)), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s created", uuid))))
	return
}

// GetAnnotations is the public API
func GetAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := isContentTypeJSON(r); err != nil {
		log.Error(err)
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
	}

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		http.Error(w, string(jsonMessage("uuid required")), http.StatusBadRequest)
		return
	}
	annotations, found, err := AnnotationsDriver.Read(uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		message := fmt.Sprintf(`{"message": "%s"}`, err.Error())
		w.Write([]byte(message))
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		message := fmt.Sprintf(`{"message":"No annotations found for content with uuid %s not found."}`, uuid)
		w.Write([]byte(message))
		return
	}
	Jason, _ := json.Marshal(annotations)
	log.Debugf("Annotations for content (uuid:%s): %s\n", Jason)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(annotations)
}
