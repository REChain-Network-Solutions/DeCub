package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/rechain/rechain/api/proto"
)

var grpcAddr string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "rechainctl",
		Short: "REChain CLI tool",
	}

	rootCmd.PersistentFlags().StringVar(&grpcAddr, "grpc-addr", "localhost:9090", "gRPC server address")

	rootCmd.AddCommand(
		nodeCmd(),
		blockCmd(),
		txCmd(),
		casCmd(),
		gossipCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func nodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Node operations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "info",
			Short: "Get node information",
			Run: func(cmd *cobra.Command, args []string) {
				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetNodeInfo(context.Background(), &proto.NodeInfoRequest{})
				if err != nil {
					log.Fatalf("Failed to get node info: %v", err)
				}

				printJSON(resp)
			},
		},
		&cobra.Command{
			Use:   "peers",
			Short: "Get connected peers",
			Run: func(cmd *cobra.Command, args []string) {
				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetPeers(context.Background(), &proto.PeersRequest{})
				if err != nil {
					log.Fatalf("Failed to get peers: %v", err)
				}

				printJSON(resp)
			},
		},
	)

	return cmd
}

func blockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block operations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "get [height]",
			Short: "Get block by height",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				height := parseUint64(args[0])

				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetBlock(context.Background(), &proto.GetBlockRequest{Height: height})
				if err != nil {
					log.Fatalf("Failed to get block: %v", err)
				}

				printJSON(resp)
			},
		},
		&cobra.Command{
			Use:   "latest",
			Short: "Get latest block",
			Run: func(cmd *cobra.Command, args []string) {
				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetLatestBlock(context.Background(), &proto.GetLatestBlockRequest{})
				if err != nil {
					log.Fatalf("Failed to get latest block: %v", err)
				}

				printJSON(resp)
			},
		},
	)

	return cmd
}

func txCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Transaction operations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "submit [type] [payload]",
			Short: "Submit a transaction",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				txType := args[0]
				payload := []byte(args[1])

				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.SubmitTx(context.Background(), &proto.SubmitTxRequest{
					Type:    txType,
					Payload: payload,
				})
				if err != nil {
					log.Fatalf("Failed to submit transaction: %v", err)
				}

				printJSON(resp)
			},
		},
		&cobra.Command{
			Use:   "get [hash]",
			Short: "Get transaction by hash",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				hash := args[0]

				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetTx(context.Background(), &proto.GetTxRequest{Hash: hash})
				if err != nil {
					log.Fatalf("Failed to get transaction: %v", err)
				}

				printJSON(resp)
			},
		},
	)

	return cmd
}

func casCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cas",
		Short: "CAS operations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "store [file]",
			Short: "Store a file in CAS",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				filePath := args[0]

				data, err := os.ReadFile(filePath)
				if err != nil {
					log.Fatalf("Failed to read file: %v", err)
				}

				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.StoreObject(context.Background(), &proto.StoreObjectRequest{
					Data:     data,
					Metadata: map[string]string{"filename": filePath},
				})
				if err != nil {
					log.Fatalf("Failed to store object: %v", err)
				}

				printJSON(resp)
			},
		},
		&cobra.Command{
			Use:   "get [cid] [output]",
			Short: "Get an object from CAS",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				cid := args[0]
				outputPath := args[1]

				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetObject(context.Background(), &proto.GetObjectRequest{Cid: cid})
				if err != nil {
					log.Fatalf("Failed to get object: %v", err)
				}

				if !resp.Found {
					log.Fatalf("Object not found")
				}

				err = os.WriteFile(outputPath, resp.Data, 0644)
				if err != nil {
					log.Fatalf("Failed to write file: %v", err)
				}

				fmt.Printf("Object saved to %s\n", outputPath)
			},
		},
	)

	return cmd
}

func gossipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gossip",
		Short: "Gossip operations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "state",
			Short: "Get gossip state",
			Run: func(cmd *cobra.Command, args []string) {
				conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				client := proto.NewRechainServiceClient(conn)
				resp, err := client.GetGossipState(context.Background(), &proto.GossipStateRequest{})
				if err != nil {
					log.Fatalf("Failed to get gossip state: %v", err)
				}

				printJSON(resp)
			},
		},
	)

	return cmd
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(data))
}

func parseUint64(s string) uint64 {
	var result uint64
	fmt.Sscanf(s, "%d", &result)
	return result
}
