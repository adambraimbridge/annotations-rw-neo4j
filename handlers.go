package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type httpHandlers struct {
	AnnotationsService annotations.Service
}

// HealthCheck does something
func (hh *httpHandlers) HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Annotation API requests",
		Name:             "Check connectivity to Neo4j --neo-url is part of the service_args in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance with at least one person loaded in it",
		Checker:          hh.Checker,
	}
}

// Checker does more stuff
//TODO use the shared utility check
func (hh *httpHandlers) Checker() (string, error) {
	err := hh.AnnotationsService.Check()
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
func (hh *httpHandlers) PutAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := isContentTypeJSON(r); err != nil {
		log.Error(err)
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	decoder := json.NewDecoder(r.Body)
	anns, err := hh.AnnotationsService.DecodeJSON(decoder)
	if err != nil {
		msg := fmt.Sprintf("Error (%v) parsing annotation request", err)
		log.Info(msg)
		writeJSONError(w, msg, http.StatusBadRequest)
		return
	}
	err = hh.AnnotationsService.Write(uuid, anns)
	if err != nil {
		msg := fmt.Sprintf("Error creating annotations (%v)", err)
		if _, ok := err.(annotations.ValidationError); ok {
			log.Error(msg)
			writeJSONError(w, msg, http.StatusBadRequest)
			return
		}
		log.Error(msg)
		writeJSONError(w, msg, http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s created", uuid))))
	return
}

// GetAnnotations returns a view of the annotations written - it is NOT the public annotations API, and
// the response format should be consistent with the PUT request body format
func (hh *httpHandlers) GetAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		writeJSONError(w, "uuid required", http.StatusBadRequest)
		return
	}
	annotations, found, err := hh.AnnotationsService.Read(uuid)
	if err != nil {
		msg := fmt.Sprintf("Error getting annotations (%v)", err)
		log.Error(msg)
		writeJSONError(w, msg, http.StatusServiceUnavailable)
		return
	}
	if !found {
		writeJSONError(w, fmt.Sprintf("No annotations found for content with uuid %s.", uuid), http.StatusNotFound)
		return
	}
	Jason, _ := json.Marshal(annotations)
	log.Debugf("Annotations for content (uuid:%s): %s\n", Jason)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(annotations)
}

// DeleteAnnotations will delete all the annotations for a piece of content
func (hh *httpHandlers) DeleteAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		writeJSONError(w, "uuid required", http.StatusBadRequest)
		return
	}
	found, err := hh.AnnotationsService.Delete(uuid)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if !found {
		writeJSONError(w, fmt.Sprintf("No annotations found for content with uuid %s.", uuid), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s deleted", uuid))))
}

func (hh *httpHandlers) CountAnnotations(w http.ResponseWriter, r *http.Request) {
	count, err := hh.AnnotationsService.Count()

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		log.Errorf("Error on read=%v\n", err)
		writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	enc := json.NewEncoder(w)

	if err := enc.Encode(count); err != nil {
		log.Errorf("Error on json encoding=%v\n", err)
		writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

}

func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}
