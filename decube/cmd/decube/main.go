package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/decube/decube/internal/api"
	"github.com/decube/decube/internal/etcd"
	"github.com/decube/decube/pkg/config"
)

var (
	configPath = flag.String("config", "./config/config.yaml", "Path to configuration file")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize etcd manager
	etcdManager := etcd.NewEtcdManager(cfg)
	if err := etcdManager.Start(); err != nil {
		log.Fatalf("Failed to start etcd: %v", err)
	}
	defer etcdManager.Stop()

	// Initialize REST API server
	var restServer *api.RESTServer
	if cfg.API.REST.Enabled {
		restServer = api.NewRESTServer(etcdManager, cfg.API.REST.Address)
		go func() {
			if err := restServer.Start(); err != nil {
				log.Printf("REST server error: %v", err)
			}
		}()
	}

	// Initialize gRPC API server
	var grpcServer *api.GRPCServer
	if cfg.API.GRPC.Enabled {
		grpcServer = api.NewGRPCServer(etcdManager)
		go func() {
			if err := grpcServer.Start(cfg.API.GRPC.Address); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	log.Printf("DeCube local control-plane started")
	log.Printf("REST API: %s", cfg.API.REST.Address)
	log.Printf("gRPC API: %s", cfg.API.GRPC.Address)
	log.Printf("etcd client: %s", cfg.Node.ListenAddress)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")

	// Stop servers
	if grpcServer != nil {
		grpcServer.Stop()
	}
	if restServer != nil {
		restServer.Stop()
	}
}
