package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jtaylor-io/safe-as-houses/api"
	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/jtaylor-io/safe-as-houses/gapi"
	"github.com/jtaylor-io/safe-as-houses/pb"
	"github.com/jtaylor-io/safe-as-houses/util"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config file", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	err = runMigrations(config.MigrationURL, config.DBSource)
	if err != nil {
		log.Fatal("could not run db migrations:", err)
	}
	log.Println("db migrations ran successfully")

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
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

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSafeAsHousesServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("starting gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server:", err)
	}
}

//go:embed doc/swagger/*
var swaggerContent embed.FS

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
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
		log.Fatal("cannot register gateway handler:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	swaggerSubDir, err := fs.Sub(swaggerContent, "doc/swagger")
	if err != nil {
		log.Fatal("cannot load swagger content:", err)
	}
	fs := http.FileServer(http.FS(swaggerSubDir))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("starting HTTP gateway server on %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway serverL:", err)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
		return
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
