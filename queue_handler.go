package main

import (
	"encoding/json"

	"github.com/Financial-Times/annotations-rw-neo4j/v4/annotations"
	"github.com/Financial-Times/annotations-rw-neo4j/v4/forwarder"

	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/kafka"
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"

	"github.com/pkg/errors"
)

// Note: this will only work for annotation messages, and not for suggestion
// because suggestions-rw-neo4j has in its config shouldConsumeMessages set to false
// and therefore the code bellow is not executed

type queueMessage struct {
	UUID        string
	Annotations annotations.Annotations
}

type queueHandler struct {
	annotationsService annotations.Service
	consumer           kafka.Consumer
	forwarder          forwarder.QueueForwarder
	originMap          map[string]string
	lifecycleMap       map[string]string
	messageType        string
	log                *logger.UPPLogger
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
			qh.log.WithMonitoringEvent("SaveNeo4j", tid, qh.messageType).WithUUID(annMsg.UUID).WithError(err).Error("Cannot write to Neo4j")
			return errors.Wrapf(err, "Failed to write message with tid=%s and uuid=%s", tid, annMsg.UUID)
		}

		qh.log.WithMonitoringEvent("SaveNeo4j", tid, qh.messageType).WithUUID(annMsg.UUID).Infof("%s successfully written in Neo4j", qh.messageType)

		//forward message to the next queue
		if qh.forwarder != nil {
			qh.log.WithTransactionID(tid).WithUUID(annMsg.UUID).Debug("Forwarding message to the next queue")
			return qh.forwarder.SendMessage(tid, originSystem, annMsg.UUID, annMsg.Annotations)
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
