package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"io"
	"time"

	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/twinj/uuid"
)

const (
	dateFormat            = "2006-01-02T03:04:05.000Z0700"
	lifecyclePropertyName = "annotationLifecycle"
)

//service def
type httpHandler struct {
	annotationsService annotations.Service
	producer           kafka.Producer
	originMap          map[string]string
	lifecycleMap       map[string]string
	messageType        string
	log                *logger.UPPLogger
}

// GetAnnotations returns a view of the annotations written - it is NOT the public annotations API, and
// the response format should be consistent with the PUT request body format
func (hh *httpHandler) GetAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]
	if uuid == "" {
		writeJSONError(w, "uuid required", http.StatusBadRequest)
		return
	}

	lifecycle := vars[lifecyclePropertyName]
	if lifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	} else if _, ok := hh.lifecycleMap[lifecycle]; !ok {
		writeJSONError(w, "annotationLifecycle not supported by this application", http.StatusBadRequest)
		return
	}

	tid := transactionidutils.GetTransactionIDFromRequest(r)
	annotations, found, err := hh.annotationsService.Read(uuid, tid, lifecycle)
	if err != nil {
		msg := fmt.Sprintf("Error getting annotations (%v)", err)
		writeJSONError(w, msg, http.StatusServiceUnavailable)
		return
	}
	if !found {
		writeJSONError(w, fmt.Sprintf("No annotations found for content with uuid %s.", uuid), http.StatusNotFound)
		return
	}
	annotationJson, _ := json.Marshal(annotations)
	hh.log.Debugf("Annotations for content (uuid:%s): %s\n", uuid, annotationJson)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(annotations)
}

// DeleteAnnotations will delete all the annotations for a piece of content
func (hh *httpHandler) DeleteAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]
	if uuid == "" {
		writeJSONError(w, "uuid required", http.StatusBadRequest)
		return
	}

	lifecycle := vars[lifecyclePropertyName]
	if lifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	} else if _, ok := hh.lifecycleMap[lifecycle]; !ok {
		writeJSONError(w, "annotationLifecycle not supported by this application", http.StatusBadRequest)
		return
	}

	tid := transactionidutils.GetTransactionIDFromRequest(r)
	found, err := hh.annotationsService.Delete(uuid, tid, lifecycle)
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

func (hh *httpHandler) CountAnnotations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lifecycle := vars[lifecyclePropertyName]
	if lifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	} else if _, ok := hh.lifecycleMap[lifecycle]; !ok {
		writeJSONError(w, "annotationLifecycle not supported by this application", http.StatusBadRequest)
		return
	}

	platformVersion, found := hh.lifecycleMap[lifecycle]
	if !found {
		writeJSONError(w, "platformVersion not found for this annotation lifecycle", http.StatusBadRequest)
		return
	}

	count, err := hh.annotationsService.Count(lifecycle, platformVersion)

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	enc := json.NewEncoder(w)

	if err := enc.Encode(count); err != nil {
		writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

// PutAnnotations handles the replacement of a set of annotations for a given bit of content
func (hh *httpHandler) PutAnnotations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := isContentTypeJSON(r); err != nil {
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	if uuid == "" {
		writeJSONError(w, "uuid required", http.StatusBadRequest)
		return
	}

	lifecycle := vars[lifecyclePropertyName]
	if lifecycle == "" {
		writeJSONError(w, "annotationLifecycle required for uuid %s"+uuid, http.StatusBadRequest)
		return
	}

	platformVersion, ok := hh.lifecycleMap[lifecycle]
	if !ok {
		writeJSONError(w, "annotationLifecycle not supported by this application", http.StatusBadRequest)
		return
	}

	anns, err := decode(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error (%v) parsing annotation request", err)
		writeJSONError(w, msg, http.StatusBadRequest)
		return
	}

	tid := transactionidutils.GetTransactionIDFromRequest(r)
	err = hh.annotationsService.Write(uuid, lifecycle, platformVersion, tid, anns)
	if err == annotations.UnsupportedPredicateErr {
		writeJSONError(w, "Please provide a valid predicate, or leave blank for the default predicate (MENTIONS)", http.StatusBadRequest)
		return
	}

	if err != nil {
		msg := fmt.Sprintf("Error creating annotations (%v)", err)
		if _, ok := err.(annotations.ValidationError); ok {
			writeJSONError(w, msg, http.StatusBadRequest)
			return
		}
		hh.log.WithMonitoringEvent("SaveNeo4j", tid, hh.messageType).WithUUID(uuid).WithError(err).Error(msg)
		writeJSONError(w, msg, http.StatusServiceUnavailable)
		return
	}
	hh.log.WithMonitoringEvent("SaveNeo4j", tid, hh.messageType).WithUUID(uuid).Infof("%s successfully written in Neo4j", hh.messageType)

	if hh.producer != nil {
		var originSystem string
		for k, v := range hh.originMap {
			if v == lifecycle {
				originSystem = k
				break
			}
		}
		if originSystem == "" {
			writeJSONError(w, "no origin-system-id could be deduced for the lifecycle parameter", http.StatusBadRequest)
			return
		}

		err = hh.forwardMessage(uuid, anns, tid, originSystem)
		if err != nil {
			msg := "Failed to forward message to queue"
			hh.log.WithTransactionID(tid).WithUUID(uuid).WithError(err).Error(msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(jsonMessage(msg)))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s created", uuid))))
	return
}

func (hh *httpHandler) forwardMessage(uuid string, anns annotations.Annotations, tid string, originSystem string) error {

	msg := map[string]interface{}{
		"uuid":         uuid,
		hh.messageType: anns,
	}

	headers := createHeader(tid, originSystem)
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	hh.log.WithTransactionID(tid).WithUUID(uuid).Info("Forwarding message to the next queue")
	return hh.producer.SendMessage(kafka.NewFTMessage(headers, string(body)))
}

func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}

func jsonMessage(msgText string) []byte {
	return []byte(fmt.Sprintf(`{"message":"%s"}`, msgText))
}

func decode(body io.Reader) (annotations.Annotations, error) {
	var anns annotations.Annotations
	err := json.NewDecoder(body).Decode(&anns)
	return anns, err
}

func isContentTypeJSON(r *http.Request) error {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "application/json") {
		return errors.New("Http Header 'Content-Type' is not 'application/json', this is a JSON API")
	}
	return nil
}

func createHeader(tid string, originSystem string) map[string]string {
	return map[string]string{
		"X-Request-Id":      tid,
		"Message-Timestamp": time.Now().Format(dateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      "concept-annotations",
		"Content-Type":      "application/json",
		"Origin-System-Id":  originSystem,
	}
}
