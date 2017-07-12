package main

import (
	"encoding/json"
	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type queueHandler struct {
	annotationsService annotations.Service
	consumer           kafka.Consumer
	producer           kafka.Producer
	originMap          map[string]string
	lifecycleMap       map[string]string
	messageType        string
}

func (qh *queueHandler) Ingest() {
	qh.consumer.StartListening(func(message kafka.FTMessage) error {

		tid, found := message.Headers[transactionidutils.TransactionIDHeader]
		if !found {
			return errors.New("Missing transaction id from message")
		}

		originSystem, found := message.Headers["Origin-System-Id"]
		if !found {
			return errors.New("Missing Origini-System-Id header from message")
		}

		lifecycle, platformVersion, err := qh.getSourceFromHeader(originSystem)
		if err != nil {
			return err
		}

		var annotationMessage map[string]interface{}
		err = json.Unmarshal([]byte(message.Body), &annotationMessage)
		if err != nil {
			return errors.Errorf("Cannot process received message %s", tid)
		}

		uuid := annotationMessage["uuid"].(string)
		if uuid == "" {
			return errors.Errorf("Cannot find `uuid` field in the received message for tid %s. Message will be ignored. ", tid)
		}

		annotationMsg := annotationMessage[qh.messageType]
		if annotationMsg == nil {
			return errors.Errorf("Cannot find `%s` field in the received message for tid %s. Message will be ignored. ", qh.messageType, tid)
		}

		err = qh.annotationsService.Write(uuid, lifecycle, platformVersion, tid, annotationMsg)
		if err != nil {
			return errors.Wrapf(err, "Failed to write message with tid=%s and uuid=%s", tid, uuid)
		}

		//forward message to the next queue
		if qh.producer != nil {
			log.WithFields(map[string]interface{}{"tid": tid, "uuid": uuid}).Info("Forwarding message to the next queue")
			return qh.producer.SendMessage(message)
		}
		return nil
	})
}

func (qh *queueHandler) getSourceFromHeader(originSystem string) (string, string, error) {
	annotationLifecycle, found := qh.originMap[originSystem]
	if !found {
		return "", "", errors.Errorf("Annotation Lifecycle not found for origin system id: %s", originSystem)
	}

	platformVersion, found := qh.lifecycleMap[annotationLifecycle]
	if !found {
		return "", "", errors.Errorf("Platform version not found for origin system id: %s and annotation lifecycle: %s", originSystem, annotationLifecycle)
	}
	return annotationLifecycle, platformVersion, nil
}
