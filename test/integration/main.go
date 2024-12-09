package main

import (
	"context"
	"log"
	"net"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"movieexample.com/gen"
	metadatatest "movieexample.com/metadata/pkg/testutil"
	movietest "movieexample.com/movie/pkg/testutil"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/memory"
	ratingtest "movieexample.com/rating/pkg/testutil"
)

const (
	metadataServiceName = "metadata"
	ratingServiceName   = "rating"
	movieServiceName    = "movie"

	metadataServiceAddr = "localhost:8081"
	ratingServiceAddr   = "localhost:8082"
	movieServiceAddr    = "localhost:8083"
)

func main() {
	log.Println("Starting integration tests...")
	ctx := context.Background()
	registry := memory.NewRegistry()
	metadataSrv := StartMetadataService(ctx, registry)
	defer metadataSrv.GracefulStop()
	ratingSrv := StartRatingService(ctx, registry)
	defer ratingSrv.GracefulStop()
	movieSrv := StartMovieService(ctx, registry)
	defer movieSrv.GracefulStop()
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.NewClient(metadataServiceAddr, opts)

	if err != nil {
		panic(err)
	}
	defer metadataConn.Close()
	ratingConn, err := grpc.NewClient(ratingServiceAddr, opts)

	if err != nil {
		panic(err)
	}
	defer ratingConn.Close()
	movieConn, err := grpc.NewClient(movieServiceAddr, opts)
	if err != nil {
		panic(err)
	}
	defer movieConn.Close()

	metadataClient := gen.NewMetadataServiceClient(metadataConn)
	ratingClient := gen.NewRatingServiceClient(ratingConn)
	movieClient := gen.NewMovieServiceClient(movieConn)

	log.Println("Saving test metadata via metadata service")
	m := &gen.Metadata{
		Id:          "the-movie",
		Title:       "The Movie",
		Description: "A great movie",
		Director:    "John Doe",
	}
	_, err = metadataClient.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: m})
	if err != nil {
		log.Fatalf("Failed to save metadata: %v", err)
	}

	log.Println("Retrieving metadata via metadata service")

	getMetadataRes, err := metadataClient.GetMetadata(ctx, &gen.GetMetadataRequest{
		MovieId: "the-movie",
	})
	if err != nil {
		log.Fatalf("Failed to retrieve metadata: %v", err)
	}
	if diff := cmp.Diff(getMetadataRes.Metadata, m, cmpopts.IgnoreUnexported(gen.Metadata{})); diff != "" {
		log.Fatalf("Metadata mismatch (-want +got):\n%s", diff)
	}
	log.Println("Getting movie details via movie service")

	log.Println("saving first rating via rating service")
	const userID = "user0"
	const recordTypeMovie = "movie"
	firstRating := int32(5)
	if _, err := ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordType:  recordTypeMovie,
		RecordId:    m.Id,
		RatingValue: firstRating,
	}); err != nil {
		log.Fatalf("Failed to save rating: %v", err)
	}

	log.Println("Retrieving aggregated rating via rating service")
	getAggregatedRatingResponse, err := ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatalf("Failed to retrieve aggregated rating: %v", err)
	}

	if got, want := getAggregatedRatingResponse.RatingValue, float64(firstRating); got != want {
		log.Fatalf("Aggregated rating mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	log.Println("Saving second rating via rating service")

	getMovieDetailsRes, err := movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{
		MovieId: m.Id,
	})
	if err != nil {
		log.Fatalf("Failed to retrieve movie details: %v", err)
	}

	wantMovieDetails := &gen.MovieDetails{
		Metadata: m,
		Rating:   getAggregatedRatingResponse.RatingValue,
	}
	if diff := cmp.Diff(getMovieDetailsRes.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("Movie details mismatch (-want +got):\n%s", diff)
	}

	secondRating := int32(1)
	if _, err := ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordType:  recordTypeMovie,
		RecordId:    m.Id,
		RatingValue: secondRating,
	},
	); err != nil {
		log.Fatalf("Failed to save rating: %v", err)
	}
	log.Println("Retrieving second aggregated rating via rating service")

	getAggregatedRatingResponse, err = ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatalf("Failed to retrieve aggregated rating: %v", err)
	}
	if got, want := getAggregatedRatingResponse.RatingValue, float64((firstRating+secondRating)/2); got != want {
		log.Fatalf("Aggregated rating mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}

	log.Println("Retrieving update movie details via movie service")
	wantMovieDetails_2 := &gen.MovieDetails{
		Metadata: m,
		Rating:   getAggregatedRatingResponse.RatingValue,
	}

	getMovieDetailsRes, err = movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{
		MovieId: m.Id,
	})
	if err != nil {
		log.Fatalf("Failed to retrieve movie details: %v", err)
	}
	if diff := cmp.Diff(getMovieDetailsRes.MovieDetails, wantMovieDetails_2, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("Movie details mismatch (-want +got):\n%s", diff)
	}

	log.Println("Integration tests passed!")

}

func StartMetadataService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting metadata service on :", metadataServiceAddr)
	h := metadatatest.NewTestMetadataGRPCServer()
	l, err := net.Listen("tcp", metadataServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddr); err != nil {
		panic(err)
	}
	return srv
}

func StartRatingService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting rating service on :", ratingServiceAddr)
	h := ratingtest.NewTestRatingGRPCServer()
	l, err := net.Listen("tcp", ratingServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddr); err != nil {
		panic(err)
	}
	return srv
}

func StartMovieService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting movie service on :", movieServiceAddr)
	h := movietest.NewTestMovieGRPCServer(registry)
	l, err := net.Listen("tcp", movieServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gen.RegisterMovieServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	id := discovery.GenerateInstanceID(movieServiceName)
	if err := registry.Register(ctx, id, movieServiceName, movieServiceAddr); err != nil {
		panic(err)
	}
	return srv

}
