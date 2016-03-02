package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/stretchr/testify/assert"
)

const knownUUID = "12345"

type test struct {
	name         string
	req          *http.Request
	dummyService annotations.Service
	statusCode   int
	contentType  string // Contents of the Content-Type header
	body         string
}

func TestPutHandler(t *testing.T) {
	assert := assert.New(t)
	body, err := ioutil.ReadFile("annotations/examplePutBody.json")
	assert.NoError(err, "Unexpected error")
	invalidBody := []byte(`{"id": "1234"}`)
	missingConceptIDBody := []byte(`"{"thing": {"prefLabel": "Apple"}`)
	tests := []test{
		{"Success", newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", body), dummyService{contentUUID: knownUUID}, http.StatusCreated, "application/json", message("Annotations for content 12345 created")},
		{"ParseError", newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", invalidBody), dummyService{contentUUID: knownUUID, failParse: true}, http.StatusBadRequest, "application/json", message("Error (TEST failing to DECODE) parsing annotation request")},
		{"ValidationError", newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", missingConceptIDBody), dummyService{contentUUID: knownUUID, failValidation: true}, http.StatusBadRequest, "application/json", message("Error creating annotations (TEST failing validation)")},
		{"NotJson", newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "text/html", body), dummyService{contentUUID: knownUUID}, http.StatusBadRequest, "application/json", message("Http Header 'Content-Type' is not 'application/json', this is a JSON API")},
		{"WriteFailed", newRequest("PUT", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", body), dummyService{contentUUID: knownUUID, failWrite: true}, http.StatusServiceUnavailable, "application/json", message("Error creating annotations (TEST failing to WRITE)")},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.dummyService}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestGetHandler(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{"Success", newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusOK, "", "[]"},
		{"NotFound", newRequest("GET", fmt.Sprintf("/content/%s/annotations", "99999"), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusNotFound, "", message("No annotations found for content with uuid 99999.")},
		{"ReadError", newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID, failRead: true}, http.StatusServiceUnavailable, "", message("Error getting annotations (TEST failing to READ)")},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.dummyService}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestDeleteHandler(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{"Success", newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusNoContent, "", message("Annotations for content 12345 deleted")},
		{"NotFound", newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", "99999"), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusNotFound, "", message("No annotations found for content with uuid 99999.")},
		{"DeleteError", newRequest("DELETE", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID, failDelete: true}, http.StatusServiceUnavailable, "", message("TEST failing to DELETE")},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.dummyService}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		if rec.Body != nil {
			assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
		}
	}
}

func TestCountHandler(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{"Success", newRequest("GET", "/content/annotations/__count", "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusOK, "", "2\n"},
		{"CountError", newRequest("GET", "/content/annotations/__count", "application/json", nil), dummyService{contentUUID: knownUUID, failCount: true}, http.StatusServiceUnavailable, "", message("TEST failing to COUNT")},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.dummyService}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.Equal(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
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

type dummyService struct {
	contentUUID    string
	failWrite      bool
	failRead       bool
	failParse      bool
	failValidation bool
	failDelete     bool
	failCount      bool
}

type dummyServiceData struct {
}

func (dS dummyService) Write(contentUUID string, thing interface{}) error {
	if dS.failValidation {
		return annotations.ValidationError{"TEST failing validation"}
	}
	if dS.failWrite {
		return errors.New("TEST failing to WRITE")
	}
	return nil
}

func (dS dummyService) Read(contentUUID string) (interface{}, bool, error) {
	if dS.failRead {
		return nil, false, errors.New("TEST failing to READ")
	}
	if contentUUID == dS.contentUUID {
		return []dummyServiceData{}, true, nil
	}
	return nil, false, nil
}

func (dS dummyService) Delete(contentUUID string) (bool, error) {
	if dS.failDelete {
		return false, errors.New("TEST failing to DELETE")
	}
	if contentUUID == dS.contentUUID {
		return true, nil
	}
	return false, nil
}

func (dS dummyService) Check() error {
	return nil
}

func (dS dummyService) DecodeJSON(*json.Decoder) (interface{}, error) {
	if dS.failParse {
		return "", errors.New("TEST failing to DECODE")
	}
	return dummyServiceData{}, nil
}

func (dS dummyService) Count() (int, error) {
	if dS.failCount {
		return 0, errors.New("TEST failing to COUNT")
	}
	return 2, nil
}

func (dS dummyService) Initialise() error {
	return nil
}

func healthHandler(http.ResponseWriter, *http.Request) {
}
