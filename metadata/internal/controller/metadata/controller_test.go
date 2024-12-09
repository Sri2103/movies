package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	gen "movieexample.com/gen/mock/metadata/repository"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

// put and get data to memory repository.
func TestControllerPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repoMock := gen.NewMockmetadataRepository(ctrl)
	c := New(repoMock)
	ctx := context.Background()
	id := "id"
	m := &model.Metadata{
		ID:          id,
		Title:       "title",
		Description: "description",
		Director:    "director",
	}

	Trctx, span := otel.Tracer("").Start(ctx, "TestControllerPut")
	defer span.End()

	repoMock.EXPECT().Put(Trctx, id, m).Return(nil)
	err := c.Put(ctx, m)
	assert.NoError(t, err)

	repoMock.EXPECT().Get(Trctx, id).Return(m, nil)
	res, err := c.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, m, res)
}

func TestController(t *testing.T) {
	tests := []struct {
		name       string
		expRepoRes *model.Metadata
		expRepoErr error
		wantRes    *model.Metadata
		wantErr    error
	}{
		{
			name:       "not found",
			expRepoErr: repository.ErrNotFound,
			wantErr:    ErrNotFound,
		},
		{
			name:       "unexpected error",
			expRepoErr: errors.New("unexpected error"),
			wantErr:    errors.New("unexpected error"),
		},
		{
			name:       "success",
			expRepoRes: &model.Metadata{},
			wantRes:    &model.Metadata{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := gen.NewMockmetadataRepository(ctrl)
			c := New(repoMock)
			ctx := context.Background()
			id := "id"
			Trctx, span := otel.Tracer("").Start(ctx, "TestControllerGet")
			repoMock.EXPECT().Get(Trctx, id).Return(tt.expRepoRes, tt.expRepoErr)

			defer span.End()
			res, err := c.Get(ctx, id)
			assert.Equal(t, tt.wantRes, res, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}
