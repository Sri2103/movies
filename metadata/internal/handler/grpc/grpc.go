package grpc

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/pkg/model"
)

// Handler is the GRPC server implementation for the Metadata service.
// It embeds the generated UnimplementedMetadataServiceServer and
// holds a reference to the metadata.Controller for handling requests.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl *metadata.Controller
}

// New creates a new GRPC server Handler that holds a reference to the
// metadata.Controller for handling requests.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

// GetMetadata is the GRPC handler for the GetMetadata RPC. It retrieves the metadata for the
// specified movie ID, or returns an error if the metadata is not found or an internal error
// occurs.
func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	ctx, span := otel.Tracer("").Start(ctx, "GetMetadata")
	defer span.End()

	meter := otel.Meter("metadata.grpc.GetMetadata")

	counter, err := meter.Int64Counter("metadata.grpc.GetMetadata.counter")
	if err != nil {
		return nil, err
	}

	counter.Add(ctx, 1, nil)

	if req == nil || req.MovieId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	span.SetAttributes(attribute.String("movie_id", req.MovieId))

	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "metadata not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "failed to get metadata")
	}

	return &gen.GetMetadataResponse{
		Metadata: model.MetadataToProto(m),
	}, nil
}

// PutMetadata is the handler for the PutMetadata RPC. It stores the metadata for the specified movie ID,
//
//	or returns an error.
func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	ctx, span := otel.Tracer("metadata").Start(ctx, "PutMetadata")
	defer span.End()

	if req == nil || req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "metadata is required")
	}

	if req.Metadata.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	span.SetAttributes(attribute.String("movie_id", req.Metadata.Id))

	m := model.MetadataFromProto(req.Metadata)

	span.SetAttributes(attribute.String("movie_id", m.ID))

	err := h.ctrl.Put(ctx, m)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to put metadata")
	}

	return &gen.PutMetadataResponse{}, nil
}
