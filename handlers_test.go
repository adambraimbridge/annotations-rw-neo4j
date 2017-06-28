package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http/httptest"
)

const knownUUID = "12345"

type HttpHandlerTestSuite struct {
	suite.Suite
	body               []byte
	annotations        annotations.Annotations
	annotationsService *mockAnnotationsService
	producer           *mockProducer
	message            kafka.FTMessage
	healthCheckHandler healthCheckHandler
}

func (suite *HttpHandlerTestSuite) SetupTest() {
	var err error
	suite.body, err = ioutil.ReadFile("annotations/examplePutBody.json")
	assert.NoError(suite.T(), err, "Unexpected error")

	suite.annotations, err = decode(bytes.NewReader(suite.body))
	assert.NoError(suite.T(), err, "Unexpected error")

	suite.annotationsService = new(mockAnnotationsService)
	suite.producer = new(mockProducer)

	headers := createHeader("tid_sample", "test_origin_system")
	msgBody, err := json.Marshal(queueMessage{knownUUID, suite.annotations})
	assert.NoError(suite.T(), err, "Unexpected error")
	suite.message = kafka.NewFTMessage(headers, string(msgBody))
	suite.healthCheckHandler = healthCheckHandler{}
}

func TestHttpHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HttpHandlerTestSuite))
}

func (suite *HttpHandlerTestSuite) TestPutHandler_Success() {
	suite.annotationsService.On("Write", knownUUID, suite.annotations).Return(nil)
	suite.producer.On("SendMessage", mock.Anything).Return(nil).Once()
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", suite.body)
	request.Header.Add("X-Request-Id", "tid_sample")
	request.Header.Add("X-Origin-System_Id", "test_origin_system")
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: true}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusCreated == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusCreated))
	assert.JSONEq(suite.T(), message("Annotations for content 12345 created"), rec.Body.String(), "Wrong body")
	suite.producer.AssertExpectations(suite.T())
}

func (suite *HttpHandlerTestSuite) TestPutHandler_ParseError() {
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", []byte(`{"id": "1234"}`))
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: false}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusBadRequest == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusCreated))
}

func (suite *HttpHandlerTestSuite) TestPutHandler_ValidationError() {
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", []byte(`"{"thing": {"prefLabel": "Apple"}`))
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: false}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusBadRequest == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusCreated))
}

func (suite *HttpHandlerTestSuite) TestPutHandler_NotJson() {
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "text/html", suite.body)
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: false}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusBadRequest == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusCreated))

}

func (suite *HttpHandlerTestSuite) TestPutHandler_WriteFailed() {
	suite.annotationsService.On("Write", knownUUID, suite.annotations).Return(errors.New("Write failed"))
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", suite.body)
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: false}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HttpHandlerTestSuite) TestPutHandler_ForwardingFailed() {
	suite.annotationsService.On("Write", knownUUID, suite.annotations).Return(nil)
	suite.producer.On("SendMessage", mock.Anything).Return(errors.New("Forwarding failed"))
	request := newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", suite.body)
	request.Header.Add("X-Request-Id", "tid_sample")
	request.Header.Add("X-Origin-System_Id", "test_origin_system")
	httpHandler := httpHandlers{AnnotationsService: suite.annotationsService, producer: suite.producer, forward: true}
	rec := httptest.NewRecorder()
	router(httpHandler, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusInternalServerError == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusInternalServerError))
}

func (suite *HttpHandlerTestSuite) TestGetHandler_Success() {
	suite.annotationsService.On("Read", knownUUID).Return(suite.annotations, true, nil)
	request := newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
	expectedResponse, err := json.Marshal(suite.annotations)
	assert.NoError(suite.T(), err, "")
	assert.JSONEq(suite.T(), string(expectedResponse), rec.Body.String(), "Wrong body")
}

func (suite *HttpHandlerTestSuite) TestGetHandler_NotFound() {
	suite.annotationsService.On("Read", knownUUID).Return(nil, false, nil)
	request := newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusNotFound == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusNotFound))
}

func (suite *HttpHandlerTestSuite) TestGetHandler_ReadError() {
	suite.annotationsService.On("Read", knownUUID).Return(nil, false, errors.New("Read error"))
	request := newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HttpHandlerTestSuite) TestDeleteHandler_Success() {
	suite.annotationsService.On("Delete", knownUUID).Return(true, nil)
	request := newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusNoContent == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusNoContent))
}

func (suite *HttpHandlerTestSuite) TestDeleteHandler_NotFound() {
	suite.annotationsService.On("Delete", knownUUID).Return(false, nil)
	request := newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusNotFound == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusNotFound))
}

func (suite *HttpHandlerTestSuite) TestDeleteHandler_DeleteError() {
	suite.annotationsService.On("Delete", knownUUID).Return(false, errors.New("Delete error"))
	request := newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

func (suite *HttpHandlerTestSuite) TestCount_Success() {
	suite.annotationsService.On("Count").Return(10, nil)
	request := newRequest("GET", "/content/annotations/__count", "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusOK == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusOK))
}

func (suite *HttpHandlerTestSuite) TestCount_CountError() {
	suite.annotationsService.On("Count").Return(0, errors.New("Count error"))
	request := newRequest("GET", "/content/annotations/__count", "application/json", nil)
	rec := httptest.NewRecorder()
	router(httpHandlers{suite.annotationsService, suite.producer, false}, suite.healthCheckHandler).ServeHTTP(rec, request)
	assert.True(suite.T(), http.StatusServiceUnavailable == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, http.StatusServiceUnavailable))
}

type QueueHandlerTestSuite struct {
	suite.Suite
	headers            map[string]string
	body               []byte
	message            kafka.FTMessage
	queueMessage       queueMessage
	annotationsService *mockAnnotationsService
	producer           *mockProducer
}

func (suite *QueueHandlerTestSuite) SetupTest() {
	var err error
	suite.headers = createHeader("tid_sample", "sample-origin")
	suite.body, err = ioutil.ReadFile("exampleAnnotationsMessage.json")
	assert.NoError(suite.T(), err, "Unexpected error")
	suite.message = kafka.NewFTMessage(suite.headers, string(suite.body))
	err = json.Unmarshal(suite.body, &suite.queueMessage)
	assert.NoError(suite.T(), err, "Unexpected error")
	suite.annotationsService = new(mockAnnotationsService)
	suite.producer = new(mockProducer)
}

func TestQueueHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(QueueHandlerTestSuite))
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest() {
	suite.annotationsService.On("Write", suite.queueMessage.UUID, suite.queueMessage.Payload).Return(nil)
	suite.producer.On("SendMessage", suite.message).Return(nil)

	qh := &queueHandler{
		AnnotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: suite.message},
		producer:           suite.producer,
		forward:            true,
	}
	qh.Ingest()

	suite.annotationsService.AssertCalled(suite.T(), "Write", suite.queueMessage.UUID, suite.queueMessage.Payload)
	suite.producer.AssertCalled(suite.T(), "SendMessage", suite.message)
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest_ForwardingDisabled() {
	suite.annotationsService.On("Write", suite.queueMessage.UUID, suite.queueMessage.Payload).Return(nil)

	qh := queueHandler{
		AnnotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: suite.message},
		producer:           suite.producer,
		forward:            false,
	}
	qh.Ingest()

	suite.annotationsService.AssertCalled(suite.T(), "Write", suite.queueMessage.UUID, suite.queueMessage.Payload)
	suite.producer.AssertNumberOfCalls(suite.T(), "SendMessage", 0)
}

func (suite *QueueHandlerTestSuite) TestQueueHandler_Ingest_JsonError() {
	body := "invalid json"
	message := kafka.NewFTMessage(suite.headers, string(body))

	qh := &queueHandler{
		AnnotationsService: suite.annotationsService,
		consumer:           mockConsumer{message: message},
		producer:           suite.producer,
		forward:            false,
	}
	qh.Ingest()

	suite.producer.AssertNumberOfCalls(suite.T(), "SendMessage", 0)
	suite.annotationsService.AssertNumberOfCalls(suite.T(), "Write", 0)
}

func newRequest(method, url, contentType string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type mockProducer struct {
	mock.Mock
}

func (mp *mockProducer) SendMessage(message kafka.FTMessage) error {
	args := mp.Called(message)
	return args.Error(0)
}

type mockConsumer struct {
	message kafka.FTMessage
	err     error
}

func (mc mockConsumer) StartListening(messageHandler func(message kafka.FTMessage) error) {
	messageHandler(mc.message)
}

func (mc mockConsumer) Shutdown() {
	return
}

func (mc mockConsumer) ConnectivityCheck() error {
	return mc.err
}

type mockAnnotationsService struct {
	mock.Mock
}

func (as *mockAnnotationsService) Write(contentUUID string, thing interface{}) (err error) {
	args := as.Called(contentUUID, thing)
	return args.Error(0)
}
func (as *mockAnnotationsService) Read(contentUUID string) (thing interface{}, found bool, err error) {
	args := as.Called(contentUUID)
	return args.Get(0), args.Bool(1), args.Error(2)
}
func (as *mockAnnotationsService) Delete(contentUUID string) (found bool, err error) {
	args := as.Called(contentUUID)
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
func (as *mockAnnotationsService) Count() (int, error) {
	args := as.Called()
	return args.Int(0), args.Error(1)
}
func (as *mockAnnotationsService) Initialise() error {
	args := as.Called()
	return args.Error(0)
}
