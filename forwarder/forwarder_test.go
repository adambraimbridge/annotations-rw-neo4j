package forwarder_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"
	"github.com/Financial-Times/annotations-rw-neo4j/v3/forwarder"

	"github.com/Financial-Times/kafka-client-go/kafka"
)

func TestSendMessage(t *testing.T) {
	p := new(mockProducer)
	f := forwarder.Forwarder{
		Producer:    p,
		MessageType: "annotations",
	}

	body, err := ioutil.ReadFile("../exampleAnnotationsMessage.json")
	if err != nil {
		t.Fatal("Unexpected error reading example message")
	}
	// The ordering of the properties is important.
	// The encoding/json package marshals maps in sorted key order and structs in the order that the fields are declared.
	message := struct {
		Annotations annotations.Annotations `json:"annotations"`
		UUID        string                  `json:"uuid"`
	}{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		t.Fatal("Unexpected error unmarshalling example message")
	}
	// Format body to be the same format as the expected output
	body, err = json.Marshal(message)
	if err != nil {
		t.Fatal("Unexpected error re-marshalling example message")
	}

	transationID := "example-transaction-id"
	originSystem := "http://cmdb.ft.com/systems/pac"
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "Create Kafka Message Headers",
			headers: nil,
		},
		{
			name: "Use Consumed Kafka Message Headers",
			headers: map[string]string{
				"X-Request-Id":      transationID,
				"Message-Timestamp": "2006-01-02T03:04:05.000Z",
				"Message-Id":        "07109c55-3870-4260-8f77-d242c1014e9f",
				"Message-Type":      "concept-annotations",
				"Content-Type":      "application/json",
				"Origin-System-Id":  originSystem,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = f.SendMessage(transationID, originSystem, test.headers, message.UUID, message.Annotations)
			if err != nil {
				t.Error("Error sending message")
			}

			res := p.getLastMessage()
			if res.Body != string(body) {
				t.Errorf("Unexpected Kafka message processed, expected: \n`%s`\n\n but recevied: \n`%s`", string(body), res.Body)
			}
			if res.Headers["X-Request-Id"] != transationID {
				t.Errorf("Unexpected Kafka X-Request-Id, expected `%s` but recevied `%s`", transationID, res.Headers["X-Request-Id"])
			}
			if res.Headers["Origin-System-Id"] != originSystem {
				t.Errorf("Unexpected Kafka Origin-System-Id, expected `%s` but recevied `%s`", originSystem, res.Headers["Origin-System-Id"])
			}
		})
	}
}

type mockProducer struct {
	message kafka.FTMessage
}

func (mp *mockProducer) SendMessage(message kafka.FTMessage) error {
	mp.message = message
	return nil
}

func (mp *mockProducer) getLastMessage() kafka.FTMessage {
	return mp.message
}

func (mp *mockProducer) ConnectivityCheck() error {
	return nil
}

func (mp *mockProducer) Shutdown() {
}
