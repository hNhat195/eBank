package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/nhat195/simple_bank/api"
	db "github.com/nhat195/simple_bank/db/sqlc"
	"github.com/nhat195/simple_bank/gapi"
	"github.com/nhat195/simple_bank/pb"
	"github.com/nhat195/simple_bank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	runDbMigrations(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	// runGinServer(config, store)

	go runGatewayServer(config, store)
	runGPCServer(config, store)

}

func runDbMigrations(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal("cannot create migration:", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("cannot migrate db:", err)
	}

	log.Println("db migration completed")
}

func runGPCServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal("cannot create listerner:", err)
	}

	log.Println("starting gRPC server on", listen.Addr().String())

	err = grpcServer.Serve(listen)

	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal("cannot register gateway server:", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", grpcMux)

	listen, err := net.Listen("tcp", config.HTTPServerAddress)

	if err != nil {
		log.Fatal("cannot create listerner:", err)
	}

	log.Println("starting HTTP server on", listen.Addr().String())

	err = http.Serve(listen, mux)

	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}

func runGinServer(config util.Config, store db.Store) {
	sever, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = sever.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
