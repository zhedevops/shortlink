package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	guid "github.com/google/uuid"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/service"
	pb "github.com/zhedevops/shortlink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ShortenerServiceServer struct {
	pb.UnimplementedShortenerServiceServer

	service *service.Service
}

type contextKey string

const userContextKey contextKey = "user"

func New(service *service.Service) *ShortenerServiceServer {
	return &ShortenerServiceServer{
		service: service,
	}
}

func (s *ShortenerServiceServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	var response pb.URLShortenResponse

	user, ok := ctx.Value(userContextKey).(model.User)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	shortys := model.NewShortys(guid.New().String(), req.GetUrl(), user.ID)
	link, err := s.service.CreateShortLink(ctx, shortys)
	if err != nil {
		return nil, err
	}
	response.SetResult(link.ShortURL)

	return &response, nil
}

func (s *ShortenerServiceServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	var response pb.URLExpandResponse

	urlStr, err := s.service.GetOriginalURL(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	response.SetResult(urlStr)

	return &response, nil
}

func (s *ShortenerServiceServer) ListUserURLs(ctx context.Context, req *empty.Empty) (*pb.UserURLsResponse, error) {
	var response pb.UserURLsResponse
	var udata []*pb.URLData

	user, ok := ctx.Value(userContextKey).(model.User)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	links, err := s.service.GetUserLinks(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		data := &pb.URLData{}
		data.SetShortUrl(link.ShortURL)
		data.SetOriginalUrl(link.OriginalURL)
		udata = append(udata, data)
	}

	response.SetUrl(udata)

	return &response, nil
}

func Serve(service *service.Service, cnf config.ServerConfig) error {
	// Нужно определить порт для сервера
	listen, err := net.Listen("tcp", cnf.GRPCAddress)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации listener: %w", err)
	}
	// Создаем gRPC сервер без зарегистрированной службы
	s := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor(service)))
	// Регистрируем сервис-хендлер
	grpcHandler := New(service)
	pb.RegisterShortenerServiceServer(s, grpcHandler)

	log.Println("сервер gRPC начал работу")
	// Получение запроса gRpc
	if err := s.Serve(listen); err != nil {
		return fmt.Errorf("ошибка при работе сервера: %w", err)
	}

	return nil
}

func unaryInterceptor(service *service.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		token := values[0]
		user, err := service.ParseAuthToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, userContextKey, user)

		return handler(ctx, req)
	}
}
