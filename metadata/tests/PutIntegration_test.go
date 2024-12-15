//go:build integration

package meadaIntegrationTests

import (
	"context"
	"log"
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"movieexample.com/gen"
	config "movieexample.com/metadata/configs"
	postgres "movieexample.com/metadata/internal/repository/postgres"
	"movieexample.com/metadata/pkg/testutil"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/memory"
)

const (
	metadataServiceAddr = "localhost:8081"
	metadataServiceName = "metadata"
)

func TestPutData(t *testing.T) {
	log.Println("Testing PutData")

	ctx := context.Background()

	registry := memory.NewRegistry()
	metadataSrv, err := StartMetadataService(ctx, registry)
	assert.NoError(t, err)
	defer metadataSrv.GracefulStop()
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.NewClient(metadataServiceAddr, opts)
	assert.NoError(t, err)
	defer metadataConn.Close()
	client := gen.NewMetadataServiceClient(metadataConn)
	mData := &gen.Metadata{
		Id:          strconv.Itoa(rand.Intn(55)),
		Title:       "The Movie",
		Description: "A great movie",
		Director:    "John Doe",
	}
	resp, err := client.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: mData})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func StartMetadataService(ctx context.Context, registry discovery.Registry) (*grpc.Server, error) {
	log.Println("Starting metadata service on :", metadataServiceAddr)
	r, err := postgres.ConnectSQL(ctx, &config.Config{})
	if err != nil {
		return nil, err
	}
	h := testutil.NewTestMetadataGRPCServer(r)
	l, err := net.Listen("tcp", metadataServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
			return
		}
	}()
	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddr); err != nil {
		panic(err)
	}
	return srv, nil
}
