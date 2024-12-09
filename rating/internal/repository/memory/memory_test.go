package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"movieexample.com/rating/pkg/model"
)

var repo = New()

func TestRepository_Get(t *testing.T) {
	type args struct {
		in0        context.Context
		recordID   model.RecordID
		recordType model.RecordType
	}
	tests := []struct {
		name    string
		r       *Repository
		args    args
		want    []model.Rating
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			r:    repo,
			args: args{
				in0:        context.Background(),
				recordID:   "1",
				recordType: model.RecordTypeMovie,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Get(tt.args.in0, tt.args.recordID, tt.args.recordType)
			assert.Equal(t, tt.want, got)
			assert.Error(t, err)
		})
	}
}

func TestRepository_Put(t *testing.T) {
	type args struct {
		in0        context.Context
		recordID   model.RecordID
		recordType model.RecordType
		rating     *model.Rating
	}
	tests := []struct {
		name    string
		r       *Repository
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			r:    repo,
			args: args{
				in0:        context.Background(),
				recordID:   "1",
				recordType: model.RecordTypeMovie,
				rating: &model.Rating{
					ID:      "1",
					UserID:  "1",
					MovieID: "1",
					Value:   1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Put(tt.args.in0, tt.args.recordID, tt.args.recordType, tt.args.rating)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
	ratings, err := repo.Get(context.Background(), "1", model.RecordTypeMovie)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ratings))
}

func TestPutToGet(t *testing.T) {
	ratingRecord := &model.Rating{
		ID:      "1",
		UserID:  "user0",
		MovieID: "1",
		Value:   1,
	}
	err := repo.Put(context.Background(), ratingRecord.MovieID, model.RecordTypeMovie, ratingRecord)
	assert.NoError(t, err)
}
