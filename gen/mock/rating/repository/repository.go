// Code generated by MockGen. DO NOT EDIT.
// Source: ./rating/internal/controller/rating/controller.go
//
// Generated by this command:
//
//	mockgen -package=repository -source=./rating/internal/controller/rating/controller.go
//

// Package repository is a generated GoMock package.
package repository

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	model "movieexample.com/rating/pkg/model"
)

// MockratingRepository is a mock of ratingRepository interface.
type MockratingRepository struct {
	ctrl     *gomock.Controller
	recorder *MockratingRepositoryMockRecorder
}

// MockratingRepositoryMockRecorder is the mock recorder for MockratingRepository.
type MockratingRepositoryMockRecorder struct {
	mock *MockratingRepository
}

// NewMockratingRepository creates a new mock instance.
func NewMockratingRepository(ctrl *gomock.Controller) *MockratingRepository {
	mock := &MockratingRepository{ctrl: ctrl}
	mock.recorder = &MockratingRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockratingRepository) EXPECT() *MockratingRepositoryMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockratingRepository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, recordID, recordType)
	ret0, _ := ret[0].([]model.Rating)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockratingRepositoryMockRecorder) Get(ctx, recordID, recordType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockratingRepository)(nil).Get), ctx, recordID, recordType)
}

// Put mocks base method.
func (m *MockratingRepository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, recordID, recordType, rating)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockratingRepositoryMockRecorder) Put(ctx, recordID, recordType, rating any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockratingRepository)(nil).Put), ctx, recordID, recordType, rating)
}

// MockrateIngester is a mock of rateIngester interface.
type MockrateIngester struct {
	ctrl     *gomock.Controller
	recorder *MockrateIngesterMockRecorder
}

// MockrateIngesterMockRecorder is the mock recorder for MockrateIngester.
type MockrateIngesterMockRecorder struct {
	mock *MockrateIngester
}

// NewMockrateIngester creates a new mock instance.
func NewMockrateIngester(ctrl *gomock.Controller) *MockrateIngester {
	mock := &MockrateIngester{ctrl: ctrl}
	mock.recorder = &MockrateIngesterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockrateIngester) EXPECT() *MockrateIngesterMockRecorder {
	return m.recorder
}

// Ingest mocks base method.
func (m *MockrateIngester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ingest", ctx)
	ret0, _ := ret[0].(chan model.RatingEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Ingest indicates an expected call of Ingest.
func (mr *MockrateIngesterMockRecorder) Ingest(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ingest", reflect.TypeOf((*MockrateIngester)(nil).Ingest), ctx)
}
