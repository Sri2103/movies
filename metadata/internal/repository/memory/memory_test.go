package memory

import (
	"context"
	"reflect"
	"testing"

	"movieexample.com/metadata/pkg/model"
)

func TestRepository_Get(t *testing.T) {
	type args struct {
		in0 context.Context
		id  string
	}
	tests := []struct {
		name    string
		r       *Repository
		args    args
		want    *model.Metadata
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Get(tt.args.in0, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_Put(t *testing.T) {
	type args struct {
		in0 context.Context
		id  string
		m   *model.Metadata
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
			r: &Repository{
				data: map[string]*model.Metadata{},
			},
			args: args{
				in0: context.Background(),
				id:  "1",
				m: &model.Metadata{
					ID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "test2",
			r: &Repository{
				data: map[string]*model.Metadata{},
			},
			args: args{
				in0: context.Background(),
				id:  "2",
				m: &model.Metadata{
					ID: "2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Put(tt.args.in0, tt.args.id, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Repository.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// test put and get
func TestRepository_PutAndGet(t *testing.T) {
	r := New()
	tests := []struct {
		name string
		r    *Repository
		m    *model.Metadata
	}{
		{
			name: "test1",
			r:    r,
			m: &model.Metadata{
				ID:          "1",
				Title:       "test-movie",
				Director:    "test-director",
				Description: "test-description",
			},
		},
		{
			name: "test2",
			r:    r,
			m: &model.Metadata{
				ID:          "2",
				Title:       "test-movie-2",
				Director:    "test-director-2",
				Description: "test-description-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Put(context.Background(), tt.m.ID, tt.m)
			if err != nil {
				t.Errorf("Repository.Put() error = %v", err)
			}
			m, err := tt.r.Get(context.Background(), tt.m.ID)
			if err != nil {
				t.Errorf("Repository.Get() error = %v", err)
			}
			if !reflect.DeepEqual(m, tt.m) {
				t.Errorf("Repository.Get() = %v, want %v", m, tt.m)
			}
		})
	}
}
