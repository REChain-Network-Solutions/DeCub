package api

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/rechain/rechain/api/proto"
)

// gRPCServer implements the Rechain gRPC service
type gRPCServer struct {
	proto.UnimplementedRechainServiceServer
	server *grpc.Server
	api    *Server
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(api *Server) *gRPCServer {
	s := grpc.NewServer()
	srv := &gRPCServer{
		server: s,
		api:    api,
	}

	proto.RegisterRechainServiceServer(s, srv)

	// Enable server reflection (for debugging)
	reflection.Register(s)

	return srv
}

// Start starts the gRPC server
func (s *gRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("gRPC server starting on %s", addr)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *gRPCServer) Stop() error {
	s.server.GracefulStop()
	return nil
}

// Implement the gRPC service methods

func (s *gRPCServer) GetNodeInfo(ctx context.Context, req *proto.NodeInfoRequest) (*proto.NodeInfoResponse, error) {
	// Get node info from REST API handler
	// This is a simplified implementation
	return &proto.NodeInfoResponse{
		NodeId:      "node-1",
		Version:     "0.1.0",
		Network:     "rechain-mainnet",
		BlockHeight: 0,
		Consensus:   "bft",
		Peers:       []string{},
		StartTime:   "2023-01-01T00:00:00Z",
	}, nil
}

func (s *gRPCServer) GetPeers(ctx context.Context, req *proto.PeersRequest) (*proto.PeersResponse, error) {
	// Simplified peer list
	peers := []*proto.Peer{
		{
			Id:       "peer-1",
			Address:  "127.0.0.1:26656",
			LastSeen: "2023-12-01T10:00:00Z",
			Connected: true,
		},
	}

	return &proto.PeersResponse{
		Peers: peers,
		Count: int32(len(peers)),
	}, nil
}

func (s *gRPCServer) GetBlock(ctx context.Context, req *proto.GetBlockRequest) (*proto.BlockResponse, error) {
	// This would call the REST API handler
	return &proto.BlockResponse{
		Block: nil,
		Found: false,
	}, nil
}

func (s *gRPCServer) GetLatestBlock(ctx context.Context, req *proto.GetLatestBlockRequest) (*proto.BlockResponse, error) {
	// This would call the REST API handler
	return &proto.BlockResponse{
		Block: nil,
		Found: false,
	}, nil
}

func (s *gRPCServer) GetBlocks(ctx context.Context, req *proto.GetBlocksRequest) (*proto.BlocksResponse, error) {
	// This would call the REST API handler
	return &proto.BlocksResponse{
		Blocks: []*proto.Block{},
		Count:  0,
	}, nil
}

func (s *gRPCServer) SubmitTx(ctx context.Context, req *proto.SubmitTxRequest) (*proto.SubmitTxResponse, error) {
	// This would call the REST API handler
	return &proto.SubmitTxResponse{
		TxId:      "tx-123",
		Status:    "submitted",
		Timestamp: "2023-12-01T10:00:00Z",
	}, nil
}

func (s *gRPCServer) GetTx(ctx context.Context, req *proto.GetTxRequest) (*proto.TxResponse, error) {
	// This would call the REST API handler
	return &proto.TxResponse{
		Tx:    nil,
		Found: false,
	}, nil
}

func (s *gRPCServer) GetTxs(ctx context.Context, req *proto.GetTxsRequest) (*proto.TxsResponse, error) {
	// This would call the REST API handler
	return &proto.TxsResponse{
		Txs:   []*proto.Transaction{},
		Count: 0,
	}, nil
}

func (s *gRPCServer) GetConsensusState(ctx context.Context, req *proto.ConsensusStateRequest) (*proto.ConsensusStateResponse, error) {
	// This would call the REST API handler
	return &proto.ConsensusStateResponse{
		Height:        0,
		Round:         0,
		Step:          "unknown",
		Proposer:      "unknown",
		Validators:    []string{"node-1"},
		MempoolSize:   0,
		LastCommitTime: "2023-12-01T10:00:00Z",
	}, nil
}

func (s *gRPCServer) StoreObject(ctx context.Context, req *proto.StoreObjectRequest) (*proto.StoreObjectResponse, error) {
	// This would call the REST API handler
	return &proto.StoreObjectResponse{
		Cid:        "cid-123",
		Size:       int64(len(req.Data)),
		Chunks:     1,
		MerkleRoot: "merkle-123",
		Uploaded:   "2023-12-01T10:00:00Z",
	}, nil
}

func (s *gRPCServer) GetObject(ctx context.Context, req *proto.GetObjectRequest) (*proto.GetObjectResponse, error) {
	// This would call the REST API handler
	return &proto.GetObjectResponse{
		Data:     []byte{},
		Metadata: map[string]string{},
		Found:    false,
	}, nil
}

func (s *gRPCServer) DeleteObject(ctx context.Context, req *proto.DeleteObjectRequest) (*proto.DeleteObjectResponse, error) {
	// This would call the REST API handler
	return &proto.DeleteObjectResponse{
		Deleted: true,
	}, nil
}

func (s *gRPCServer) ListObjects(ctx context.Context, req *proto.ListObjectsRequest) (*proto.ListObjectsResponse, error) {
	// This would call the REST API handler
	return &proto.ListObjectsResponse{
		Objects: []*proto.ObjectInfo{},
		Count:   0,
	}, nil
}

func (s *gRPCServer) GetGossipState(ctx context.Context, req *proto.GossipStateRequest) (*proto.GossipStateResponse, error) {
	// This would call the REST API handler
	return &proto.GossipStateResponse{
		State:      map[string]string{},
		PeerCount:  0,
		LastSync:   "2023-12-01T10:00:00Z",
	}, nil
}

func (s *gRPCServer) UpdateGossipState(ctx context.Context, req *proto.UpdateGossipStateRequest) (*proto.UpdateGossipStateResponse, error) {
	// This would call the REST API handler
	return &proto.UpdateGossipStateResponse{
		Updated: true,
	}, nil
}

func (s *gRPCServer) QueryGossip(ctx context.Context, req *proto.QueryGossipRequest) (*proto.QueryGossipResponse, error) {
	// This would call the REST API handler
	return &proto.QueryGossipResponse{
		Value: "",
		Found: false,
	}, nil
}
