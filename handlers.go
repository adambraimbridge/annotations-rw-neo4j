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
}

type queueHandler struct {
	AnnotationsService annotations.Service
	consumer           kafka.Consumer
	producer           kafka.Producer
}

// annotationsMessage represents a message ingested from the queue
type queueMessage struct {
	UUID    string                  `json:"uuid,omitempty"`
	Payload annotations.Annotations `json:"annotations,omitempty"`
}

var originMap = map[string]string{
	"concept-suggestor":                            "annotations-v2",
	"http://cmdb.ft.com/systems/methode-web-pub":   "annotations-v1",
	"http://cmdb.ft.com/systems/next-video-editor": "annotations-next-video",
}

var lifecycleMap = map[string]string{
	"annotations-v1":         "v1",
	"annotations-v2":         "v2",
	"annotations-next-video": "next-video",
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
	originSystem := r.Header.Get("X-Origin-System-Id")
	if "" == originSystem {
		err := errors.New("Missing X-Origini-System-Id header from message")
		log.Error(err)
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
		return
	}

	annotationLifecycle, platformVersion, err := getSourceFromHeader(originSystem)
	if err != nil {
		log.Error(err)
		http.Error(w, string(jsonMessage(err.Error())), http.StatusBadRequest)
		return
	}

	anns, err := decode(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error (%v) parsing annotation request", err)
		log.Info(msg)
		writeJSONError(w, msg, http.StatusBadRequest)
		return
	}
	err = hh.AnnotationsService.Write(uuid, annotationLifecycle, platformVersion, anns)
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
	if hh.producer != nil {
		err = hh.forwardMessage(queueMessage{uuid, anns}, tid, originSystem)
		if err != nil {
			msg := "Failed to forward message to queue"
			log.WithFields(map[string]interface{}{"tid": tid, "uuid": uuid, "error": err.Error()}).Error(msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(jsonMessage(msg)))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jsonMessage(fmt.Sprintf("Annotations for content %s created", uuid))))
	return
}

func (hh *httpHandlers) forwardMessage(queueMessage queueMessage, tid string, originSystem string) error {
	headers := createHeader(tid, originSystem)
	body, err := json.Marshal(queueMessage)
	if err != nil {
		return err
	}
	return hh.producer.SendMessage(kafka.NewFTMessage(headers, string(body)))
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

	annotationLifecycle := vars["annotationLifecycle"]
	if annotationLifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	}

	annotations, found, err := hh.AnnotationsService.Read(uuid, annotationLifecycle)
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

	annotationLifecycle := vars["annotationLifecycle"]
	if annotationLifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	}

	found, err := hh.AnnotationsService.Delete(uuid, annotationLifecycle)
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
	vars := mux.Vars(r)
	annotationLifecycle := vars["annotationLifecycle"]
	if annotationLifecycle == "" {
		writeJSONError(w, "annotationLifecycle required", http.StatusBadRequest)
		return
	}

	platformVersion, found := lifecycleMap[annotationLifecycle]
	if !found {
		writeJSONError(w, "platformVersion not found for this annotation lifecycle", http.StatusBadRequest)
		return
	}

	count, err := hh.AnnotationsService.Count(annotationLifecycle, platformVersion)

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

		tid, found := message.Headers[transactionidutils.TransactionIDHeader]
		if !found {
			return errors.New("Missing transaction id from message")
		}

		originSystem, found := message.Headers["Origin-System-Id"]
		if !found {
			return errors.New("Missing Origini-System-Id header from message")
		}

		annotationLifecycle, platformVersion, err := getSourceFromHeader(originSystem)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(message.Body), &annotationMessage)
		if err != nil {
			return errors.Errorf("Cannot process received message %s", tid)
		}

		err = qh.AnnotationsService.Write(annotationMessage.UUID, annotationLifecycle, platformVersion, annotationMessage.Payload)
		if err != nil {
			return errors.Wrapf(err, "Failed to write message with tid=%s and uuid=%s", tid, annotationMessage.UUID)
		}

		//forward message to next queue
		if qh.producer != nil {
			log.WithFields(map[string]interface{}{"tid": tid, "uuid": annotationMessage.UUID}).Info("Forwarding message to next queue")
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

func decode(body io.Reader) (annotations.Annotations, error) {
	var anns annotations.Annotations
	err := json.NewDecoder(body).Decode(&anns)
	return anns, err
}

func getSourceFromHeader(originSystem string) (string, string, error) {
	annotationLifecycle, found := originMap[originSystem]
	if !found {
		return "", "", errors.Errorf("Annotation Lifecycle not found for origin system id: %s", originSystem)
	}

	platformVersion, found := lifecycleMap[annotationLifecycle]
	if !found {
		return "", "", errors.Errorf("Platform version not found for origin system id: %s and annotation lifecycle: %s", originSystem, annotationLifecycle)
	}
	return annotationLifecycle, platformVersion, nil
}
