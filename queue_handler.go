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
}

//Note: this will only work for annotation messages, and not for suggestion
type queueMessage struct {
	UUID        string
	Annotations annotations.Annotations
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

		annMsg := new(queueMessage)
		err = json.Unmarshal([]byte(message.Body), &annMsg)
		if err != nil {
			return errors.Errorf("Cannot process received message %s", tid)
		}

		err = qh.annotationsService.Write(annMsg.UUID, lifecycle, platformVersion, tid, annMsg.Annotations)
		if err != nil {
			return errors.Wrapf(err, "Failed to write message with tid=%s and uuid=%s", tid, annMsg.UUID)
		}

		//forward message to the next queue
		if qh.producer != nil {
			log.WithFields(map[string]interface{}{"tid": tid, "uuid": annMsg.UUID}).Info("Forwarding message to the next queue")
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
