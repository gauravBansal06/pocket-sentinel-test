package endpoint

import (
	"context"

	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	pb "github.com/LambdatestIncPrivate/protobuf/golang/bookkeeping/host/v1"
	"github.com/go-kit/kit/endpoint"
	kep "github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type HostEndpoints struct {
	GetHost    kep.Endpoint
	CreateHost kep.Endpoint
}

func NewHostEndpoints(svc service.HostService, logger lumber.Logger) HostEndpoints {
	return HostEndpoints{
		GetHost:    makeGetEndpoint(svc, logger),
		CreateHost: makeSetEndpoint(svc, logger),
	}
}

func makeGetEndpoint(s service.HostService, logger lumber.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			logger.Infof("Received headers: %+v", md)
		}
		req := request.(service.GetByIDRequest)
		host, err := s.GetByID(ctx, req.ID)
		if err != nil {
			logger.Errorf("Unable to get host by ID: %s", req.ID)
			return pb.HostServiceGetResponse{Host: nil}, status.Errorf(codes.NotFound,
				err.Error())
		}
		// send response headers
		logger.Infof("Sending response header")
		header := metadata.Pairs("header-key", "val")
		grpc.SendHeader(ctx, header)
		return service.GetByIDResponse{Host: host}, nil
	}
}

func makeSetEndpoint(s service.HostService, logger lumber.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			logger.Infof("Received headers: %+v", md)
		}
		req := request.(service.CreateRequest)
		id, hash, err := s.Create(ctx, req.Host)
		if err != nil {
			logger.Errorf("Unable to create host using payload: %+v", req.Host)
			return pb.HostServiceGetResponse{Host: nil}, status.Errorf(codes.Internal,
				err.Error())
		}
		logger.Infof("Sending response header")
		header := metadata.Pairs("header-key", "val")
		grpc.SendHeader(ctx, header)

		logger.Debugf("Host created successfully with id: %s and hash: %d", id, hash)
		return service.CreateResponse{UUID: id, Hash: hash}, nil
	}
}
