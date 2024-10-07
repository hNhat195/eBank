package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/nhat195/simple_bank/api"
	db "github.com/nhat195/simple_bank/db/sqlc"
	"github.com/nhat195/simple_bank/gapi"
	"github.com/nhat195/simple_bank/pb"
	"github.com/nhat195/simple_bank/util"
	"github.com/nhat195/simple_bank/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDbMigrations(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	// runGinServer(config, store)

	go runTaskProcessor(redisOpt, store)
	go runGatewayServer(config, store, taskDistributor)
	runGPCServer(config, store, taskDistributor)

}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().Msg("task processor started")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start task processor")
	}
}

func runDbMigrations(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration:")
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("cannot migrate db:")
	}

	log.Info().Msg("db migration completed")
}

func runGPCServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listerner")
	}

	log.Info().Msgf("Starting gRPC server on %s", listen.Addr().String())

	err = grpcServer.Serve(listen)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}

}

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot register gateway server:")
	}

	mux := http.NewServeMux()

	mux.Handle("/", grpcMux)

	listen, err := net.Listen("tcp", config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listerner:")
	}

	log.Info().Msgf("starting HTTP server on %s", listen.Addr().String())

	handler := gapi.HttpLogger(mux)

	err = http.Serve(listen, handler)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server:")
	}

}

func runGinServer(config util.Config, store db.Store) {
	sever, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	err = sever.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server:")
	}

}
