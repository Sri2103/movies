package movie

import (
	"context"
	"reflect"
	"testing"

	"movieexample.com/movie/pkg/model"
)

func TestController_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		c       *Controller
		args    args
		want    *model.MovieDetails
		wantErr bool
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
