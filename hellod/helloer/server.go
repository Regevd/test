package main

import (
	pb "github.com/traiana/okro/hellod/api/hello/v1"
	"golang.org/x/net/context"
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func (s *server) Hi(ctx context.Context, in *pb.HiRequest) (*pb.HiResponse, error) {
	log.Infof("saying hi to %s", in.Name)
	return &pb.HiResponse{Message: "hi " + in.Name}, nil
}

func (s *server) Bye(ctx context.Context, in *pb.ByeRequest) (*pb.ByeResponse, error) {
	log.Infof("saying bye to %s", in.Name)
	return &pb.ByeResponse{Message: "bye " + in.Name}, nil
}
