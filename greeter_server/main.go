package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/yangxikun/go-grpc-client-side-lb-example/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	port = ":50051"
)

var stuckDuration time.Duration

type healthServer struct{}

func (h *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	log.Println("recv health check for service:", req.Service)
	if stuckDuration == time.Second {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (h *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	log.Println("recv health watch for service:", req.Service)
	resp := new(grpc_health_v1.HealthCheckResponse)
	if stuckDuration == time.Second {
		resp.Status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	} else {
		resp.Status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	for range time.NewTicker(time.Second).C {
		err := stream.Send(resp)
		if err != nil {
			return status.Error(codes.Canceled, "Stream has ended.")
		}
	}
	return nil
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	time.Sleep(stuckDuration)
	return &pb.HelloReply{Message: "Hello " + in.Name + "! From " + GetIP()}, nil
}

func main() {
	// simulate busy server
	stuckDuration = time.Duration(rand.NewSource(time.Now().UnixNano()).Int63()%2) * time.Second

	if stuckDuration == time.Second {
		log.Println("I will stuck one second!!!")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	grpc_health_v1.RegisterHealthServer(s, &healthServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func GetIP() string {
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip.String() != "127.0.0.1" {
				return ip.String()
			}
		}
	}
	return ""
}
