package api

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/decube/decube/api/proto"
	"github.com/decube/decube/internal/etcd"
)

// GRPCServer provides gRPC API endpoints for the DeCube control-plane
type GRPCServer struct {
	proto.UnimplementedDeCubeServiceServer
	etcdManager *etcd.EtcdManager
	server      *grpc.Server
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(etcdManager *etcd.EtcdManager) *GRPCServer {
	s := grpc.NewServer()
	srv := &GRPCServer{
		etcdManager: etcdManager,
		server:      s,
	}

	proto.RegisterDeCubeServiceServer(s, srv)

	// Enable server reflection (for debugging)
	reflection.Register(s)

	return srv
}

// Start starts the gRPC server
func (s *GRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("gRPC server starting on %s", addr)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop() error {
	s.server.GracefulStop()
	return nil
}

// Pod operations
func (s *GRPCServer) CreatePod(ctx context.Context, req *proto.CreatePodRequest) (*proto.CreatePodResponse, error) {
	// Convert to internal format and store
	pod := req.Pod
	key := fmt.Sprintf("/pods/%s/%s", pod.Namespace, pod.Name)

	// Store pod data
	podData := map[string]interface{}{
		"name":       pod.Name,
		"namespace":  pod.Namespace,
		"status":     pod.Status,
		"node_name":  pod.NodeName,
		"created_at": pod.CreatedAt,
		"updated_at": pod.UpdatedAt,
		"labels":     pod.Labels,
		"annotations": pod.Annotations,
	}

	// Serialize and store
	data, _ := json.Marshal(podData)
	err := s.etcdManager.Put(ctx, key, string(data))
	if err != nil {
		return &proto.CreatePodResponse{
			Pod:     nil,
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.CreatePodResponse{
		Pod:     pod,
		Success: true,
		Error:   "",
	}, nil
}

func (s *GRPCServer) GetPod(ctx context.Context, req *proto.GetPodRequest) (*proto.GetPodResponse, error) {
	key := fmt.Sprintf("/pods/%s/%s", req.Namespace, req.Name)

	data, err := s.etcdManager.Get(ctx, key)
	if err != nil {
		return &proto.GetPodResponse{
			Pod:   nil,
			Found: false,
			Error: err.Error(),
		}, nil
	}

	// Parse pod data
	var podData map[string]interface{}
	json.Unmarshal([]byte(data), &podData)

	pod := &proto.Pod{
		Name:        getString(podData, "name"),
		Namespace:   getString(podData, "namespace"),
		Status:      getString(podData, "status"),
		NodeName:    getString(podData, "node_name"),
		CreatedAt:   getString(podData, "created_at"),
		UpdatedAt:   getString(podData, "updated_at"),
		Labels:      getStringMap(podData, "labels"),
		Annotations: getStringMap(podData, "annotations"),
	}

	return &proto.GetPodResponse{
		Pod:   pod,
		Found: true,
		Error: "",
	}, nil
}

func (s *GRPCServer) ListPods(ctx context.Context, req *proto.ListPodsRequest) (*proto.ListPodsResponse, error) {
	prefix := fmt.Sprintf("/pods/%s/", req.Namespace)
	podsMap, err := s.etcdManager.GetWithPrefix(ctx, prefix)
	if err != nil {
		return &proto.ListPodsResponse{
			Pods:  nil,
			Count: 0,
		}, err
	}

	var pods []*proto.Pod
	for _, data := range podsMap {
		var podData map[string]interface{}
		json.Unmarshal([]byte(data), &podData)

		pod := &proto.Pod{
			Name:        getString(podData, "name"),
			Namespace:   getString(podData, "namespace"),
			Status:      getString(podData, "status"),
			NodeName:    getString(podData, "node_name"),
			CreatedAt:   getString(podData, "created_at"),
			UpdatedAt:   getString(podData, "updated_at"),
			Labels:      getStringMap(podData, "labels"),
			Annotations: getStringMap(podData, "annotations"),
		}
		pods = append(pods, pod)
	}

	return &proto.ListPodsResponse{
		Pods:  pods,
		Count: int32(len(pods)),
	}, nil
}

func (s *GRPCServer) UpdatePod(ctx context.Context, req *proto.UpdatePodRequest) (*proto.UpdatePodResponse, error) {
	pod := req.Pod
	key := fmt.Sprintf("/pods/%s/%s", pod.Namespace, pod.Name)

	// Check if pod exists
	_, err := s.etcdManager.Get(ctx, key)
	if err != nil {
		return &proto.UpdatePodResponse{
			Pod:     nil,
			Success: false,
			Error:   "Pod not found",
		}, nil
	}

	// Update pod data
	podData := map[string]interface{}{
		"name":       pod.Name,
		"namespace":  pod.Namespace,
		"status":     pod.Status,
		"node_name":  pod.NodeName,
		"created_at": pod.CreatedAt,
		"updated_at": pod.UpdatedAt,
		"labels":     pod.Labels,
		"annotations": pod.Annotations,
	}

	data, _ := json.Marshal(podData)
	err = s.etcdManager.Put(ctx, key, string(data))
	if err != nil {
		return &proto.UpdatePodResponse{
			Pod:     nil,
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.UpdatePodResponse{
		Pod:     pod,
		Success: true,
		Error:   "",
	}, nil
}

func (s *GRPCServer) DeletePod(ctx context.Context, req *proto.DeletePodRequest) (*proto.DeletePodResponse, error) {
	key := fmt.Sprintf("/pods/%s/%s", req.Namespace, req.Name)

	err := s.etcdManager.Delete(ctx, key)
	if err != nil {
		return &proto.DeletePodResponse{
			Deleted: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeletePodResponse{
		Deleted: true,
		Error:   "",
	}, nil
}

// Snapshot operations
func (s *GRPCServer) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	// Create snapshot
	snapshotData, err := s.etcdManager.CreateSnapshot(ctx)
	if err != nil {
		return &proto.CreateSnapshotResponse{
			Snapshot: nil,
			Success:  false,
			Error:    err.Error(),
		}, nil
	}

	snapshot := &proto.Snapshot{
		Id:           fmt.Sprintf("snap-%d", time.Now().Unix()),
		Name:         req.Name,
		Status:       "completed",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		SizeBytes:    int64(len(snapshotData)),
		EtcdRevision: "unknown",
		Checksum:     "unknown",
		Metadata:     req.Metadata,
	}

	// Store snapshot metadata
	snapData := map[string]interface{}{
		"id":            snapshot.Id,
		"name":          snapshot.Name,
		"status":        snapshot.Status,
		"created_at":    snapshot.CreatedAt,
		"size_bytes":    snapshot.SizeBytes,
		"etcd_revision": snapshot.EtcdRevision,
		"checksum":      snapshot.Checksum,
		"metadata":      snapshot.Metadata,
	}

	data, _ := json.Marshal(snapData)
	key := fmt.Sprintf("/snapshots/%s", snapshot.Id)
	err = s.etcdManager.Put(ctx, key, string(data))
	if err != nil {
		return &proto.CreateSnapshotResponse{
			Snapshot: nil,
			Success:  false,
			Error:    err.Error(),
		}, nil
	}

	return &proto.CreateSnapshotResponse{
		Snapshot: snapshot,
		Success:  true,
		Error:    "",
	}, nil
}

func (s *GRPCServer) GetSnapshot(ctx context.Context, req *proto.GetSnapshotRequest) (*proto.GetSnapshotResponse, error) {
	key := fmt.Sprintf("/snapshots/%s", req.Id)

	data, err := s.etcdManager.Get(ctx, key)
	if err != nil {
		return &proto.GetSnapshotResponse{
			Snapshot: nil,
			Found:    false,
			Error:    err.Error(),
		}, nil
	}

	var snapData map[string]interface{}
	json.Unmarshal([]byte(data), &snapData)

	snapshot := &proto.Snapshot{
		Id:           getString(snapData, "id"),
		Name:         getString(snapData, "name"),
		Status:       getString(snapData, "status"),
		CreatedAt:    getString(snapData, "created_at"),
		SizeBytes:    getInt64(snapData, "size_bytes"),
		EtcdRevision: getString(snapData, "etcd_revision"),
		Checksum:     getString(snapData, "checksum"),
		Metadata:     getStringMap(snapData, "metadata"),
	}

	return &proto.GetSnapshotResponse{
		Snapshot: snapshot,
		Found:    true,
		Error:    "",
	}, nil
}

func (s *GRPCServer) ListSnapshots(ctx context.Context, req *proto.ListSnapshotsRequest) (*proto.ListSnapshotsResponse, error) {
	prefix := "/snapshots/"
	snapshotsMap, err := s.etcdManager.GetWithPrefix(ctx, prefix)
	if err != nil {
		return &proto.ListSnapshotsResponse{
			Snapshots: nil,
			Count:     0,
		}, err
	}

	var snapshots []*proto.Snapshot
	for _, data := range snapshotsMap {
		var snapData map[string]interface{}
		json.Unmarshal([]byte(data), &snapData)

		snapshot := &proto.Snapshot{
			Id:           getString(snapData, "id"),
			Name:         getString(snapData, "name"),
			Status:       getString(snapData, "status"),
			CreatedAt:    getString(snapData, "created_at"),
			SizeBytes:    getInt64(snapData, "size_bytes"),
			EtcdRevision: getString(snapData, "etcd_revision"),
			Checksum:     getString(snapData, "checksum"),
			Metadata:     getStringMap(snapData, "metadata"),
		}
		snapshots = append(snapshots, snapshot)
	}

	return &proto.ListSnapshotsResponse{
		Snapshots: snapshots,
		Count:     int32(len(snapshots)),
	}, nil
}

func (s *GRPCServer) RestoreSnapshot(ctx context.Context, req *proto.RestoreSnapshotRequest) (*proto.RestoreSnapshotResponse, error) {
	// Simplified implementation
	return &proto.RestoreSnapshotResponse{
		Success:          false,
		Error:            "Snapshot restore not implemented",
		RestoredRevision: "",
	}, nil
}

func (s *GRPCServer) DeleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {
	key := fmt.Sprintf("/snapshots/%s", req.Id)

	err := s.etcdManager.Delete(ctx, key)
	if err != nil {
		return &proto.DeleteSnapshotResponse{
			Deleted: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeleteSnapshotResponse{
		Deleted: true,
		Error:   "",
	}, nil
}

// Lease operations
func (s *GRPCServer) CreateLease(ctx context.Context, req *proto.CreateLeaseRequest) (*proto.CreateLeaseResponse, error) {
	lease := &proto.Lease{
		Id:         fmt.Sprintf("lease-%d", time.Now().Unix()),
		Holder:     req.Holder,
		TtlSeconds: req.TtlSeconds,
		GrantedAt:  time.Now().UTC().Format(time.RFC3339),
		ExpiresAt:  time.Now().Add(time.Duration(req.TtlSeconds) * time.Second).UTC().Format(time.RFC3339),
		Metadata:   req.Metadata,
	}

	// Store lease data
	leaseData := map[string]interface{}{
		"id":          lease.Id,
		"holder":      lease.Holder,
		"ttl_seconds": lease.TtlSeconds,
		"granted_at":  lease.GrantedAt,
		"expires_at":  lease.ExpiresAt,
		"metadata":    lease.Metadata,
	}

	data, _ := json.Marshal(leaseData)
	key := fmt.Sprintf("/leases/%s", lease.Id)
	err := s.etcdManager.Put(ctx, key, string(data))
	if err != nil {
		return &proto.CreateLeaseResponse{
			Lease:   nil,
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.CreateLeaseResponse{
		Lease:   lease,
		Success: true,
		Error:   "",
	}, nil
}

func (s *GRPCServer) GetLease(ctx context.Context, req *proto.GetLeaseRequest) (*proto.GetLeaseResponse, error) {
	key := fmt.Sprintf("/leases/%s", req.Id)

	data, err := s.etcdManager.Get(ctx, key)
	if err != nil {
		return &proto.GetLeaseResponse{
			Lease: nil,
			Found: false,
			Error: err.Error(),
		}, nil
	}

	var leaseData map[string]interface{}
	json.Unmarshal([]byte(data), &leaseData)

	lease := &proto.Lease{
		Id:         getString(leaseData, "id"),
		Holder:     getString(leaseData, "holder"),
		TtlSeconds: getInt64(leaseData, "ttl_seconds"),
		GrantedAt:  getString(leaseData, "granted_at"),
		ExpiresAt:  getString(leaseData, "expires_at"),
		Metadata:   getStringMap(leaseData, "metadata"),
	}

	return &proto.GetLeaseResponse{
		Lease: lease,
		Found: true,
		Error: "",
	}, nil
}

func (s *GRPCServer) ListLeases(ctx context.Context, req *proto.ListLeasesRequest) (*proto.ListLeasesResponse, error) {
	prefix := "/leases/"
	leasesMap, err := s.etcdManager.GetWithPrefix(ctx, prefix)
	if err != nil {
		return &proto.ListLeasesResponse{
			Leases: nil,
			Count:  0,
		}, err
	}

	var leases []*proto.Lease
	for _, data := range leasesMap {
		var leaseData map[string]interface{}
		json.Unmarshal([]byte(data), &leaseData)

		lease := &proto.Lease{
			Id:         getString(leaseData, "id"),
			Holder:     getString(leaseData, "holder"),
			TtlSeconds: getInt64(leaseData, "ttl_seconds"),
			GrantedAt:  getString(leaseData, "granted_at"),
			ExpiresAt:  getString(leaseData, "expires_at"),
			Metadata:   getStringMap(leaseData, "metadata"),
		}
		leases = append(leases, lease)
	}

	return &proto.ListLeasesResponse{
		Leases: leases,
		Count:  int32(len(leases)),
	}, nil
}

func (s *GRPCServer) RenewLease(ctx context.Context, req *proto.RenewLeaseRequest) (*proto.RenewLeaseResponse, error) {
	key := fmt.Sprintf("/leases/%s", req.Id)

	// Get existing lease
	data, err := s.etcdManager.Get(ctx, key)
	if err != nil {
		return &proto.RenewLeaseResponse{
			Lease:   nil,
			Success: false,
			Error:   "Lease not found",
		}, nil
	}

	var leaseData map[string]interface{}
	json.Unmarshal([]byte(data), &leaseData)

	// Update lease
	ttl := req.TtlSeconds
	if ttl == 0 {
		ttl = getInt64(leaseData, "ttl_seconds")
	}

	now := time.Now().UTC()
	leaseData["ttl_seconds"] = ttl
	leaseData["granted_at"] = now.Format(time.RFC3339)
	leaseData["expires_at"] = now.Add(time.Duration(ttl) * time.Second).Format(time.RFC3339)

	updatedData, _ := json.Marshal(leaseData)
	err = s.etcdManager.Put(ctx, key, string(updatedData))
	if err != nil {
		return &proto.RenewLeaseResponse{
			Lease:   nil,
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	lease := &proto.Lease{
		Id:         getString(leaseData, "id"),
		Holder:     getString(leaseData, "holder"),
		TtlSeconds: getInt64(leaseData, "ttl_seconds"),
		GrantedAt:  getString(leaseData, "granted_at"),
		ExpiresAt:  getString(leaseData, "expires_at"),
		Metadata:   getStringMap(leaseData, "metadata"),
	}

	return &proto.RenewLeaseResponse{
		Lease:   lease,
		Success: true,
		Error:   "",
	}, nil
}

func (s *GRPCServer) DeleteLease(ctx context.Context, req *proto.DeleteLeaseRequest) (*proto.DeleteLeaseResponse, error) {
	key := fmt.Sprintf("/leases/%s", req.Id)

	err := s.etcdManager.Delete(ctx, key)
	if err != nil {
		return &proto.DeleteLeaseResponse{
			Deleted: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeleteLeaseResponse{
		Deleted: true,
		Error:   "",
	}, nil
}

// Replication operations
func (s *GRPCServer) ReplicateState(ctx context.Context, req *proto.ReplicateStateRequest) (*proto.ReplicateStateResponse, error) {
	// Apply state entries to local etcd
	for _, entry := range req.Entries {
		key := string(entry.Key)
		value := string(entry.Value)
		err := s.etcdManager.Put(ctx, key, value)
		if err != nil {
			return &proto.ReplicateStateResponse{
				Success: false,
				Error:   err.Error(),
				AppliedRevision: 0,
			}, nil
		}
	}

	return &proto.ReplicateStateResponse{
		Success: true,
		Error:   "",
		AppliedRevision: 0, // Would get actual revision
	}, nil
}

func (s *GRPCServer) GetReplicationStatus(ctx context.Context, req *proto.GetReplicationStatusRequest) (*proto.GetReplicationStatusResponse, error) {
	// Simplified implementation
	peers := []*proto.PeerStatus{
		{
			Address:     "127.0.0.1:2380",
			Connected:   true,
			LastHeartbeat: time.Now().Unix(),
			Revision:    0,
		},
	}

	return &proto.GetReplicationStatusResponse{
		Peers:         peers,
		IsLeader:      s.etcdManager.IsLeader(),
		LeaderAddress: s.etcdManager.GetLeaderAddr(),
		CurrentRevision: 0,
	}, nil
}

// Helper functions
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key]; ok {
		if num, ok := val.(float64); ok {
			return int64(num)
		}
	}
	return 0
}

func getStringMap(data map[string]interface{}, key string) map[string]string {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
			return result
		}
	}
	return nil
}
