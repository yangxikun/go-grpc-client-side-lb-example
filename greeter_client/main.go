package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/yangxikun/go-grpc-client-side-lb-example/pb"
	_ "github.com/yangxikun/go-grpc-client-side-lb-example/resolver/dns"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/resolver"
)

const (
	defaultName = "rokety"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	var address string
	var timeout int
	flag.IntVar(&timeout, "timeout", 1, "greet rpc call timeout")
	flag.StringVar(&address, "address", "localhost:50051", "grpc server addr")
	flag.Parse()
	// Set up a connection to the server.
	resolver.SetDefaultScheme("custom_dns")
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s","MethodConfig": [{"Name": [{"Service": "helloworld.Greeter"}], "RetryPolicy": {"MaxAttempts":2, "InitialBackoff": "0.1s", "MaxBackoff": "1s", "BackoffMultiplier": 2.0, "RetryableStatusCodes": ["UNAVAILABLE", "CANCELLED"]}}], "HealthCheckConfig": {"ServiceName": "helloworld.Greeter"}}`, roundrobin.Name)),
		grpc.WithBlock(), grpc.WithBackoffMaxDelay(time.Second))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	for range time.Tick(time.Second) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
		if err != nil {
			log.Printf("could not greet: %v\n", err)
		} else {
			log.Printf("Greeting: %s", r.Message)
		}
		cancel()
	}
}
