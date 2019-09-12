package main

import (
	"encoding/json"
	"testing"

	"io/ioutil"

	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type QueueHandlerTestSuite struct {
	suite.Suite
	headers            map[string]string
	body               []byte
	message            kafka.FTMessage
	queueMessage       queueMessage
	annotationsService *mockAnnotationsService
	producer           *mockProducer
	originMap          map[string]string
	lifecycleMap       map[string]string
	tid                string
}

func (suite *QueueHandlerTestSuite) SetupTest() {
	var err error
	logger.InitDefaultLogger("annotations-rw")
	suite.tid = "tid_sample"
	suite.headers = createHeader(suite.tid, "http://cmdb.ft.com/systems/methode-web-pub")
	suite.body, err = ioutil.ReadFile("exampleAnnotationsMessage.json")
	assert.NoError(suite.T(), err, "Unexpected error")
	suite.message = kafka.NewFTMessage(suite.headers, string(suite.body))
	err = json.Unmarshal(suite.body, &suite.queueMessage)
	assert.NoError(suite.T(), err, "Unexpected error")
	suite.annotationsService = new(mockAnnotationsService)
	suite.producer = new(mockProducer)

	suite.originMap, suite.lifecycleMap, _, err = readConfigMap("annotation-config.json")
	assert.NoError(suite.T(), err, "Unexpected config error")
}

func TestQueueHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(QueueHandlerTestSuite))
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest() {
	suite.annotationsService.On("Write", suite.queueMessage.UUID, annotationLifecycle, platformVersion, suite.tid, suite.queueMessage.Annotations).Return(nil)
	suite.producer.On("SendMessage", suite.message).Return(nil)

	qh := &queueHandler{
		annotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: suite.message},
		producer:           suite.producer,
		originMap:          suite.originMap,
		lifecycleMap:       suite.lifecycleMap,
	}
	qh.Ingest()

	suite.annotationsService.AssertCalled(suite.T(), "Write", suite.queueMessage.UUID, annotationLifecycle, platformVersion, suite.tid, suite.queueMessage.Annotations)
	suite.producer.AssertCalled(suite.T(), "SendMessage", suite.message)
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest_ProducerNil() {
	suite.annotationsService.On("Write", suite.queueMessage.UUID, annotationLifecycle, platformVersion, suite.tid, suite.queueMessage.Annotations).Return(nil)

	qh := queueHandler{
		annotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: suite.message},
		producer:           nil,
		originMap:          suite.originMap,
		lifecycleMap:       suite.lifecycleMap,
	}
	qh.Ingest()

	suite.annotationsService.AssertCalled(suite.T(), "Write", suite.queueMessage.UUID, annotationLifecycle, platformVersion, suite.tid, suite.queueMessage.Annotations)
	suite.producer.AssertNumberOfCalls(suite.T(), "SendMessage", 0)
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest_JsonError() {
	body := "invalid json"
	message := kafka.NewFTMessage(suite.headers, string(body))

	qh := &queueHandler{
		annotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: message},
		producer:           suite.producer,
		originMap:          suite.originMap,
		lifecycleMap:       suite.lifecycleMap,
	}
	qh.Ingest()

	suite.producer.AssertNumberOfCalls(suite.T(), "SendMessage", 0)
	suite.annotationsService.AssertNumberOfCalls(suite.T(), "Write", 0)
}
