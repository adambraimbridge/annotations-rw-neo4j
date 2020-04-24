package main

import (
	"encoding/json"

	"github.com/Financial-Times/annotations-rw-neo4j/v4/annotations"

	"github.com/Financial-Times/kafka-client-go/kafka"

	"github.com/stretchr/testify/mock"
)

type mockForwarder struct {
	mock.Mock
}

func (mf *mockForwarder) SendMessage(transactionID string, originSystem string, uuid string, annotations annotations.Annotations) error {
	args := mf.Called(transactionID, originSystem, uuid, annotations)
	return args.Error(0)
}

type mockAnnotationsService struct {
	mock.Mock
}

func (as *mockAnnotationsService) Write(contentUUID string, annotationLifecycle string, platformVersion string, tid string, thing interface{}) (err error) {
	args := as.Called(contentUUID, annotationLifecycle, platformVersion, tid, thing)
	return args.Error(0)
}
func (as *mockAnnotationsService) Read(contentUUID string, tid string, annotationLifecycle string) (thing interface{}, found bool, err error) {
	args := as.Called(contentUUID, tid, annotationLifecycle)
	return args.Get(0), args.Bool(1), args.Error(2)
}
func (as *mockAnnotationsService) Delete(contentUUID string, tid string, annotationLifecycle string) (found bool, err error) {
	args := as.Called(contentUUID, tid, annotationLifecycle)
	return args.Bool(0), args.Error(1)
}
func (as *mockAnnotationsService) Check() (err error) {
	args := as.Called()
	return args.Error(0)
}
func (as *mockAnnotationsService) DecodeJSON(decoder *json.Decoder) (thing interface{}, err error) {
	args := as.Called(decoder)
	return args.Get(0), args.Error(1)
}
func (as *mockAnnotationsService) Count(annotationLifecycle string, platformVersion string) (int, error) {
	args := as.Called(annotationLifecycle, platformVersion)
	return args.Int(0), args.Error(1)
}
func (as *mockAnnotationsService) Initialise() error {
	args := as.Called()
	return args.Error(0)
}

type mockConsumer struct {
	message kafka.FTMessage
	err     error
}

func (mc mockConsumer) StartListening(messageHandler func(message kafka.FTMessage) error) {
	_ = messageHandler(mc.message)
}

func (mc mockConsumer) Shutdown() {
}

func (mc mockConsumer) ConnectivityCheck() error {
	return mc.err
}
