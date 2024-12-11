package movie_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	gen "movieexample.com/gen/mock/movie/repository"
	modelMetadata "movieexample.com/metadata/pkg/model"
	"movieexample.com/movie/internal/controller/movie"
	ratingModel "movieexample.com/rating/pkg/model"
)

func TestGetMovieDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metaGatewayMock := gen.NewMockmetadataGateway(ctrl)
	ratingGatewayMock := gen.NewMockratingGateway(ctrl)

	movieController := movie.New(ratingGatewayMock, metaGatewayMock)
	ctx := context.Background()
	id := "id"
	metaGatewayMock.EXPECT().Get(ctx, id).Return(&modelMetadata.Metadata{}, nil)

	ratingGatewayMock.EXPECT().GetAggregatedRating(ctx, ratingModel.RecordID(id), ratingModel.RecordTypeMovie).Return(float64(5.0), nil)

	md, err := movieController.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, float64(5.0), *md.Rating)
}
