package controller_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	gen "movieexample.com/gen/mock/rating/repository"
	"movieexample.com/rating/internal/controller/rating"
	"movieexample.com/rating/pkg/model"
)

func TestControllerPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repoMock := gen.NewMockratingRepository(ctrl)
	c := rating.NewController(repoMock, nil)

	ctx := context.Background()
	id := "id"
	rating := model.Rating{
		UserID: model.UserID(id),
		Value:  5,
	}
	recordType := model.RecordTypeMovie
	repoMock.EXPECT().Put(ctx, model.RecordID(id), recordType, &rating).Return(nil)
	err := c.PutRating(ctx, model.RecordID(id), recordType, &rating)
	assert.NoError(t, err)
}
