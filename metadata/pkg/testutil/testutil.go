package testutil

import (
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	"movieexample.com/metadata/internal/repository/memory"
)

// NewTestMetadataGRPCServer creates a new metadata gRPC server to be used in tests.
func NewTestMetadataGRPCServer(repo metadata.Repository) gen.MetadataServiceServer {
	if repo == nil {
		repo = memory.New()
	}
	r := repo
	ctrl := metadata.New(r)

	return grpchandler.New(ctrl)
}
