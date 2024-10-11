package grpc

import (
	"context"
	"github.com/Azimkhan/hw12_13_14_15_calendar/gen/events/pb"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/grpc/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	grpcServer    *grpc.Server
	lsn           net.Listener
	eventsService *service.EventsService
}

func (s *Server) Serve() error {
	return s.grpcServer.Serve(s.lsn)
}

func (s *Server) Stop() {
	s.grpcServer.Stop()
}

func (s *Server) CreateGatewayMux(ctx context.Context) (*grpc.Server, error) {
	gwmux := runtime.NewServeMux()
	err := pb.RegisterEventServiceHandlerServer(ctx, gwmux, s.eventsService)
	if err != nil {
		return nil, err
	}
	return s.grpcServer, nil
}
func NewServer(calendar *app.App, conf *conf.GRPCConf) (*Server, error) {
	eventsService := service.NewEventsService(calendar)

	// gRPC server
	lsn, err := net.Listen("tcp", conf.BindAddr)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryLoggingInterceptor(calendar.Logger),
		),
	)
	pb.RegisterEventServiceServer(grpcServer, eventsService)
	reflection.Register(grpcServer)

	return &Server{
		grpcServer:    grpcServer,
		lsn:           lsn,
		eventsService: eventsService,
	}, nil
}
