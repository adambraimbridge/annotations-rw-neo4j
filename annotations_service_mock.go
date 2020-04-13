package main

import (
	"encoding/json"

	"github.com/stretchr/testify/mock"
)

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
