package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/metadata/pkg/model"
)

var (
	memoryRepo   = memory.New()
	metadataCtrl = metadata.New(memoryRepo)
	handler      = New(metadataCtrl)
)

func TestHandler_GetMetadata(t *testing.T) {
	memoryRepo.Put(context.Background(), "1", &model.Metadata{
		ID:          "1",
		Title:       "title1",
		Description: "description1",
		Director:    "director1",
	})
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *Handler
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			h:    handler,
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/metadata?id=1", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.GetMetadata(tt.args.w, tt.args.r)
			t.Log(tt.args.w.(*httptest.ResponseRecorder).Result().StatusCode)
			t.Log(tt.args.w.(*httptest.ResponseRecorder).Body.String())
		})
	}
}

func TestHandler_PutMetadata(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *Handler
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			h:    handler,
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPut, "/metadata", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model.Metadata{
				ID:          "1",
				Title:       "title1",
				Description: "description1",
				Director:    "director1",
			}

			jsonData, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			tt.args.r.Body = io.NopCloser(bytes.NewReader(jsonData))
			tt.h.PutMetadata(tt.args.w, tt.args.r)

			t.Log(tt.args.w.(*httptest.ResponseRecorder).Result().StatusCode)

			d, err := memoryRepo.Get(context.Background(), "1")
			if err != nil {
				t.Error(err)
			}
			if d.Title != m.Title {
				t.Error("title not match")
			}
		})
	}
}
