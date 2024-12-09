package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	gen "movieexample.com/gen/mock/metadata/repository"
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
	repoMock.EXPECT().Put(ctx, id, m).Return(nil)
	err := c.Put(ctx, m)
	assert.NoError(t, err)

	repoMock.EXPECT().Get(ctx, id).Return(m, nil)
	res, err := c.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, m, res)
}
