package main

import (
    "log"
    "fmt"
    "net"
    "time"

    rls "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
    "github.com/juju/ratelimit"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    //"google.golang.org/grpc/reflection"
)

// server is used to implement rls.RateLimitService
type server struct {
    bucket *ratelimit.Bucket
}

func (s *server) ShouldRateLimit(ctx context.Context, request *rls.RateLimitRequest) (*rls.RateLimitResponse, error) {
    // logic to rate limit every second request
    var overallCode rls.RateLimitResponse_Code
    fmt.Println(request.Domain)
    fmt.Println(request.Descriptors)
    if s.bucket.TakeAvailable(1) == 0 {
        overallCode = rls.RateLimitResponse_OVER_LIMIT
    } else {
        overallCode = rls.RateLimitResponse_OK
    }

    response := &rls.RateLimitResponse{OverallCode: overallCode}
    return response, nil
}

func main() {
    // create a TCP listener on port 8089
    lis, err := net.Listen("tcp", ":8081")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    log.Printf("listening on %s", lis.Addr())

    // create a gRPC server and register the RateLimitService server
    s := grpc.NewServer()

    rls.RegisterRateLimitServiceServer(s, &server{
        bucket: ratelimit.NewBucket(60 *time.Second, 1),
    })
    //reflection.Register(s)

    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
