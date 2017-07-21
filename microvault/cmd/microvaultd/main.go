package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ratelimitkit "github.com/go-kit/kit/ratelimit"

	"context"

	"github.com/juju/ratelimit"
	"github.com/thomasdarimont/gopb/microvault"
	"github.com/thomasdarimont/gopb/microvault/pb"
	"google.golang.org/grpc"
)

// run with
// ./microvaultd -http=:18080 -grpc=:18081

// test with
// curl -v -d '{"password":"bubu"}' http://localhost:18080/hash
// curl -v -d '{"password":"bubu","hash":"insert_hash_here"}' http://localhost:18080/validate

func main() {
	var (
		httpAddr = flag.String("http", ":8080", "http listen address")
		gRPCAddr = flag.String("grpc", ":8081", "gRPC listen address")
	)
	flag.Parse()
	ctx := context.Background()
	srv := microvault.NewService()
	errChan := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	rlbucket := ratelimit.NewBucket(1*time.Second, 5)
	hashEndpoint := microvault.MakeHashEndpoint(srv)
	{
		hashEndpoint = ratelimitkit.NewTokenBucketThrottler(rlbucket, time.Sleep)(hashEndpoint)
	}
	validateEndpoint := microvault.MakeValidateEndpoint(srv)
	{
		validateEndpoint = ratelimitkit.NewTokenBucketThrottler(rlbucket, time.Sleep)(validateEndpoint)
	}
	endpoints := microvault.Endpoints{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}

	// HTTP transport
	go func() {
		log.Println("http:", *httpAddr)
		handler := microvault.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()

	// gRPC transport
	go func() {
		listener, err := net.Listen("tcp", *gRPCAddr)
		if err != nil {
			errChan <- err
			return
		}
		log.Println("grpc:", *gRPCAddr)
		handler := microvault.NewGRPCServer(endpoints)
		gRPCServer := grpc.NewServer()
		pb.RegisterMicroVaultServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()

	log.Fatalln(<-errChan)
}
