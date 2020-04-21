package forwarder_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/annotations-rw-neo4j/v3/annotations"
	"github.com/Financial-Times/annotations-rw-neo4j/v3/forwarder"

	"github.com/Financial-Times/kafka-client-go/kafka"
)

type InputMessage struct {
	Annotations annotations.Annotations `json:"annotations"`
	UUID        string                  `json:"uuid"`
}

func TestSendMessage(t *testing.T) {
	const transationID = "example-transaction-id"
	const originSystem = "http://cmdb.ft.com/systems/pac"
	const expectedAnnotationsOutputBody = `{"payload":{"annotations":[{"thing":{"id":"http://api.ft.com/things/2cca9e2a-2248-3e48-abc1-93d718b91bbe","prefLabel":"China Politics \u0026 Policy","types":["http://www.ft.com/ontology/Topic"],"predicate":"majorMentions"},"provenances":[{"scores":[{"scoringSystem":"http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM","value":1},{"scoringSystem":"http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM","value":1}]}]}],"uuid":"3a636e78-5a47-11e7-9bc8-8055f264aa8b"},"contentUri":"http://annotations-rw-neo4j.svc.ft.com/content/3a636e78-5a47-11e7-9bc8-8055f264aa8b","lastModified":"%s"}`
	const expectedSuggestionsOutputBody = `{"payload":{"Suggestions":[{"thing":{"id":"http://api.ft.com/things/2cca9e2a-2248-3e48-abc1-93d718b91bbe","prefLabel":"China Politics \u0026 Policy","types":["http://www.ft.com/ontology/Topic"],"predicate":"majorMentions"},"provenances":[{"scores":[{"scoringSystem":"http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM","value":1},{"scoringSystem":"http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM","value":1}]}]}],"uuid":"3a636e78-5a47-11e7-9bc8-8055f264aa8b"},"contentUri":"http://suggestions-rw-neo4j.svc.ft.com/content/3a636e78-5a47-11e7-9bc8-8055f264aa8b","lastModified":"%s"}`

	body, err := ioutil.ReadFile("../exampleAnnotationsMessage.json")
	if err != nil {
		t.Fatal("Unexpected error reading example message")
	}
	inputMessage := InputMessage{}
	err = json.Unmarshal(body, &inputMessage)
	if err != nil {
		t.Fatal("Unexpected error unmarshalling example message")
	}
	tests := []struct {
		name         string
		headers      map[string]string
		messageType  string
		expectedBody string
	}{
		{
			name:         "Annotations Message without Headers",
			headers:      nil,
			messageType:  "annotations",
			expectedBody: expectedAnnotationsOutputBody,
		},
		{
			name:         "Suggestions Message without Headers",
			headers:      nil,
			messageType:  "Suggestions",
			expectedBody: expectedSuggestionsOutputBody,
		},
		{
			name: "Use Incoming Kafka Message Headers",
			headers: map[string]string{
				"X-Request-Id":      transationID,
				"Message-Timestamp": "2006-01-02T03:04:05.000Z",
				"Message-Id":        "07109c55-3870-4260-8f77-d242c1014e9f",
				"Message-Type":      "concept-annotations",
				"Content-Type":      "application/json",
				"Origin-System-Id":  originSystem,
			},
			messageType:  "annotations",
			expectedBody: expectedAnnotationsOutputBody,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := new(mockProducer)
			f := forwarder.Forwarder{
				Producer:    p,
				MessageType: test.messageType,
			}

			err = f.SendMessage(transationID, originSystem, test.headers, inputMessage.UUID, inputMessage.Annotations)
			if err != nil {
				t.Error("Error sending message")
			}

			res := p.getLastMessage()
			if res.Body != fmt.Sprintf(test.expectedBody, res.Headers["Message-Timestamp"]) {
				t.Errorf("Unexpected Kafka message processed, expected: \n`%s`\n\n but recevied: \n`%s`", test.expectedBody, res.Body)
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
