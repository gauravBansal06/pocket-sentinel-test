package grpc

import (
	"context"
	"net"
	"sync"

	"github.com/LambdatestIncPrivate/exemplar/config"
	"github.com/LambdatestIncPrivate/exemplar/pkg/endpoint"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	pb "github.com/LambdatestIncPrivate/protobuf/golang/bookkeeping/host/v1"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(config *config.Config, ctx context.Context, wg *sync.WaitGroup, db *sqlx.DB, logger lumber.Logger) error {
	defer wg.Done()

	svcLogger := logger.WithFields(lumber.Fields{"service": "model-service"})

	var (
		service    = service.NewHostService(svcLogger, db)
		endpoints  = endpoint.NewHostEndpoints(service, logger)
		grpcServer = NewGRPCServer(endpoints, logger, []grpctransport.ServerOption{})
	)

	//  start grpc listener
	grpcListener, err := net.Listen("tcp", ":"+config.GRPCPort)
	if err != nil {
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
		return err
	}

	server := grpc.NewServer()
	pb.RegisterHostServiceServer(server, grpcServer)
	// Register reflection service on gRPC server.
	reflection.Register(server)

	errChan := make(chan error)

	go func() {
		logger.Infof("Starting grpc server on port %s", config.GRPCPort)
		// service connections
		if err := server.Serve(grpcListener); err != nil {
			logger.Errorf("listen: %#v", err)
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Infof("Caller has requested graceful shutdown. shutting down the grpc server")
		if err := grpcListener.Close(); err != nil {
			logger.Errorf("grpc server Shutdown:", "error", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}
