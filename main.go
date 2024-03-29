package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtaylor-io/safe-as-houses/api"
	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/jtaylor-io/safe-as-houses/gapi"
	"github.com/jtaylor-io/safe-as-houses/mail"
	"github.com/jtaylor-io/safe-as-houses/pb"
	"github.com/jtaylor-io/safe-as-houses/util"
	"github.com/jtaylor-io/safe-as-houses/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config file")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	err = runMigrations(config.MigrationURL, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("could not run db migrations: " + config.MigrationURL)
	}
	log.Info().Msg("db migrations ran successfully")

	store := db.NewStore(connPool)
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runTaskProcessor(config, redisOpt, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}

func runMigrations(migrationUrl string, dbSource string) error {
	m, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		return fmt.Errorf("cannot create new migration instance: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("cannot run migrate up: %w", err)
	}
	return nil
}

func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSafeAsHousesServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Info().Msgf("starting gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server")
	}
}

//go:embed doc/swagger/*
var swaggerContent embed.FS

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	jsonOptions := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOptions)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSafeAsHousesHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register gateway handler")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	swaggerSubDir, err := fs.Sub(swaggerContent, "doc/swagger")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load swagger content")
	}
	fs := http.FileServer(http.FS(swaggerSubDir))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Info().Msgf("starting HTTP gateway server on %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
		return
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
