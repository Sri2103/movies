package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/metadata/pkg/model"
)

var (
	memoryStore = memory.New()
	controller  = metadata.New(memoryStore)
	handler     = New(controller)
)

func TestHandler_GetMetadata(t *testing.T) {
	type args struct {
		ctx context.Context
		req *gen.GetMetadataRequest
	}

	tests := []struct {
		name    string
		h       *Handler
		args    args
		want    *gen.GetMetadataResponse
		wantErr bool
	}{
		{
			name: "test-1",
			h:    handler,
			args: args{
				ctx: context.Background(),
				req: &gen.GetMetadataRequest{
					MovieId: "1",
				},
			},
			want: &gen.GetMetadataResponse{
				Metadata: &gen.Metadata{
					Id:          "1",
					Title:       "The Matrix",
					Director:    "The Wachowskis",
					Description: "A computer hacker.",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// put data into memory store
			if err := memoryStore.Put(context.Background(), "1", &model.Metadata{
				ID:          "1",
				Title:       "The Matrix",
				Director:    "The Wachowskis",
				Description: "A computer hacker.",
			},
			); err != nil {
				t.Errorf("memoryStore.Put() error = %v", err)
			}

			got, err := tt.h.GetMetadata(tt.args.ctx, tt.args.req)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			t.Log(got, "got metadata")
		})
	}
}

func TestHandler_PutMetadata(t *testing.T) {
	type args struct {
		ctx context.Context
		req *gen.PutMetadataRequest
	}

	tests := []struct {
		name    string
		h       *Handler
		args    args
		want    *gen.PutMetadataResponse
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PutMetadata(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.PutMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.PutMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
