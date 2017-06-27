package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/twinj/uuid"
	"io"
	"time"
)

type httpHandlers struct {
	AnnotationsService annotations.Service
	producer           kafka.Producer
	forward            bool
}

type queueHandler struct {
	AnnotationsService annotations.Service
	consumer           kafka.Consumer
	producer           kafka.Producer
	forward            bool
}

// annotationsMessage represents a message ingested from the queue
type queueMessage struct {
	UUID    string      `json:"uuid,omitempty"`
	Payload interface{} `json:"annotations,omitempty"`
}

const dateFormat = "2006-01-02T03:04:05.000Z0700"

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
	anns, err := decode(r.Body)
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

	tid := transactionidutils.GetTransactionIDFromRequest(r)
	originSystem := r.Header.Get("X-Origin-System-Id")
	if hh.producer != nil && hh.forward {
		err = hh.forwardMessage(uuid, tid, originSystem, anns)
		if err != nil {
			msg := "Failed to forward message to queue"
			log.WithFields(map[string]interface{}{"tid": tid, "uuid": uuid, "error": err}).Error(msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(jsonMessage(msg)))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s created", uuid))))
	return
}

func (hh *httpHandlers) forwardMessage(u string, tid string, originSystem string, annotations interface{}) error {
	headers := createHeader(tid, originSystem)
	body, err := json.Marshal(queueMessage{u, annotations})
	if err != nil {
		return err
	}
	message := kafka.NewFTMessage(headers, string(body))
	return hh.producer.SendMessage(message)
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

func (qh queueHandler) Ingest() {
	qh.consumer.StartListening(func(message kafka.FTMessage) error {
		annotationMessage := new(queueMessage)

		tid, found := message.Headers["X-Request-Id"]
		if !found {
			return errors.New("Missing transaction id from message")
		}

		log.WithFields(map[string]interface{}{"tid": tid}).Infof("Received write request %s from queue", tid)
		err := json.Unmarshal([]byte(message.Body), &annotationMessage)
		if err != nil {
			return errors.Wrapf(err, "Cannot read message body for %s", tid)
		}

		//write received message to neo4j
		err = qh.AnnotationsService.Write(annotationMessage.UUID, annotationMessage.Payload)
		if err != nil {
			return errors.Wrapf(err, "Failed to write message with tid=%s and uuid=%s", tid, annotationMessage.UUID)
		}

		//forward message to next queue
		if qh.producer != nil && qh.forward {
			return qh.producer.SendMessage(message)
		}
		return nil
	})
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

func decode(body io.Reader) ([]annotations.Annotation, error) {
	var anns []annotations.Annotation
	err := json.NewDecoder(body).Decode(&anns)
	return anns, err
}
