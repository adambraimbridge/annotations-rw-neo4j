package forwarder

import (
	"encoding/json"
	"time"

	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"
	"github.com/Financial-Times/kafka-client-go/kafka"

	uuid "github.com/satori/go.uuid"
)

type QueueForwarder interface {
	SendMessage(transactionID string, originSystem string, headers map[string]string, uuid string, annotations annotations.Annotations) error
}

type Forwarder struct {
	Producer    kafka.Producer
	MessageType string
}

func (f Forwarder) SendMessage(transactionID string, originSystem string, headers map[string]string, uuid string, annotations annotations.Annotations) error {
	if headers == nil {
		headers = f.CreateHeaders(transactionID, originSystem)
	}
	body, err := f.marshalAnnotations(uuid, annotations)
	if err != nil {
		return err
	}

	return f.Producer.SendMessage(kafka.NewFTMessage(headers, body))
}

func (f Forwarder) CreateHeaders(transactionID string, originSystem string) map[string]string {
	const dateFormat = "2006-01-02T03:04:05.000Z0700"
	messageUUID, _ := uuid.NewV4()
	return map[string]string{
		"X-Request-Id":      transactionID,
		"Message-Timestamp": time.Now().Format(dateFormat),
		"Message-Id":        messageUUID.String(),
		"Message-Type":      "concept-annotations",
		"Content-Type":      "application/json",
		"Origin-System-Id":  originSystem,
	}
}

func (f Forwarder) marshalAnnotations(uuid string, annotations annotations.Annotations) (string, error) {
	msg := map[string]interface{}{
		"uuid":        uuid,
		f.MessageType: annotations,
	}

	// Given the type of data we are marshalling, there is no possible input that can trigger an error here
	// but we are handling errors just to be principled
	body, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
