package grpc

// Guide for metadata handling in grpc
// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

import (
	"context"
	"fmt"

	"github.com/LambdatestIncPrivate/exemplar/pkg/endpoint"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/models"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	pb "github.com/LambdatestIncPrivate/protobuf/golang/bookkeeping/host/v1"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type grpcServer struct {
	pb.UnimplementedHostServiceServer
	getHost    grpctransport.Handler
	createHost grpctransport.Handler
}

func NewGRPCServer(endpoints endpoint.HostEndpoints, logger lumber.Logger, serverOptions []grpctransport.ServerOption) pb.HostServiceServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	options = append(options, serverOptions...)

	return &grpcServer{
		getHost: grpctransport.NewServer(
			endpoints.GetHost,
			decodeGRPCgetHostRequest,
			encodeGRPCgetHostResponse,
			options...,
		),
		createHost: grpctransport.NewServer(
			endpoints.CreateHost,
			decodeGRPCCreateHostRequest,
			encodeGRPCCreateHostResponse,
			options...,
		),
	}
}

func (s *grpcServer) Get(ctx context.Context, req *pb.HostServiceGetRequest) (*pb.HostServiceGetResponse, error) {
	_, rep, err := s.getHost.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.HostServiceGetResponse), nil
}

func (s *grpcServer) Set(ctx context.Context, req *pb.HostServiceSetRequest) (*pb.HostServiceSetResponse, error) {
	_, rep, err := s.createHost.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.HostServiceSetResponse), nil
}

// For this handler we are directly expecting pb implemented structs
func decodeGRPCgetHostRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.HostServiceGetRequest)
	return service.GetByIDRequest{ID: req.Uuid}, nil
}

func decodeGRPCCreateHostRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.HostServiceSetRequest)
	host, err := models.HostFromDto(req.Host)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert dto to host %s", err)
	}
	return service.CreateRequest{Host: host}, nil
}

func encodeGRPCgetHostResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(service.GetByIDResponse)
	hostDto, err := reply.Host.ToDto()
	if err != nil {
		return nil, fmt.Errorf("Unable to convert model to dto %s", err)
	}
	return &pb.HostServiceGetResponse{Host: hostDto}, nil
}

func encodeGRPCCreateHostResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(service.CreateResponse)
	return &pb.HostServiceSetResponse{Uuid: reply.UUID, Hash: reply.Hash}, nil
}
