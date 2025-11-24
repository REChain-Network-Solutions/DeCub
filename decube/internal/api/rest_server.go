package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/decube/decube/internal/etcd"
)

// RESTServer provides REST API endpoints for the DeCube control-plane
type RESTServer struct {
	etcdManager *etcd.EtcdManager
	router      *mux.Router
	server      *http.Server
}

// NewRESTServer creates a new REST server
func NewRESTServer(etcdManager *etcd.EtcdManager, address string) *RESTServer {
	rs := &RESTServer{
		etcdManager: etcdManager,
		router:      mux.NewRouter(),
	}

	rs.setupRoutes()

	rs.server = &http.Server{
		Addr:         address,
		Handler:      rs.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return rs
}

// Start starts the REST server
func (rs *RESTServer) Start() error {
	log.Printf("Starting REST server on %s", rs.server.Addr)
	return rs.server.ListenAndServe()
}

// Stop stops the REST server
func (rs *RESTServer) Stop() error {
	return rs.server.Close()
}

// setupRoutes sets up the API routes
func (rs *RESTServer) setupRoutes() {
	api := rs.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	rs.router.HandleFunc("/health", rs.healthHandler).Methods("GET")

	// Pods
	api.HandleFunc("/pods", rs.listPodsHandler).Methods("GET")
	api.HandleFunc("/pods", rs.createPodHandler).Methods("POST")
	api.HandleFunc("/pods/{name}", rs.getPodHandler).Methods("GET")
	api.HandleFunc("/pods/{name}", rs.updatePodHandler).Methods("PUT")
	api.HandleFunc("/pods/{name}", rs.deletePodHandler).Methods("DELETE")

	// Snapshots
	api.HandleFunc("/snapshots", rs.listSnapshotsHandler).Methods("GET")
	api.HandleFunc("/snapshots", rs.createSnapshotHandler).Methods("POST")
	api.HandleFunc("/snapshots/{id}", rs.getSnapshotHandler).Methods("GET")
	api.HandleFunc("/snapshots/{id}/restore", rs.restoreSnapshotHandler).Methods("POST")
	api.HandleFunc("/snapshots/{id}", rs.deleteSnapshotHandler).Methods("DELETE")

	// Leases
	api.HandleFunc("/leases", rs.listLeasesHandler).Methods("GET")
	api.HandleFunc("/leases", rs.createLeaseHandler).Methods("POST")
	api.HandleFunc("/leases/{id}", rs.getLeaseHandler).Methods("GET")
	api.HandleFunc("/leases/{id}/renew", rs.renewLeaseHandler).Methods("POST")
	api.HandleFunc("/leases/{id}", rs.deleteLeaseHandler).Methods("DELETE")

	// Node info
	rs.router.HandleFunc("/node/info", rs.nodeInfoHandler).Methods("GET")
}

// healthHandler handles health check requests
func (rs *RESTServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"is_leader": rs.etcdManager.IsLeader(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Pod handlers
func (rs *RESTServer) listPodsHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	prefix := fmt.Sprintf("/pods/%s/", namespace)
	pods, err := rs.etcdManager.GetWithPrefix(r.Context(), prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var podList []map[string]interface{}
	for key, value := range pods {
		var pod map[string]interface{}
		if err := json.Unmarshal([]byte(value), &pod); err != nil {
			continue
		}
		podList = append(podList, pod)
	}

	response := map[string]interface{}{
		"pods":  podList,
		"count": len(podList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) createPodHandler(w http.ResponseWriter, r *http.Request) {
	var pod map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set timestamps
	now := time.Now().UTC().Format(time.RFC3339)
	pod["created_at"] = now
	pod["updated_at"] = now

	// Default namespace
	if pod["namespace"] == nil {
		pod["namespace"] = "default"
	}

	name, ok := pod["name"].(string)
	if !ok {
		http.Error(w, "Pod name is required", http.StatusBadRequest)
		return
	}

	namespace, _ := pod["namespace"].(string)
	key := fmt.Sprintf("/pods/%s/%s", namespace, name)

	podJSON, _ := json.Marshal(pod)
	err := rs.etcdManager.Put(r.Context(), key, string(podJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pod":     pod,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) getPodHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	key := fmt.Sprintf("/pods/%s/%s", namespace, name)
	podJSON, err := rs.etcdManager.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	var pod map[string]interface{}
	if err := json.Unmarshal([]byte(podJSON), &pod); err != nil {
		http.Error(w, "Invalid pod data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pod":   pod,
		"found": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) updatePodHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	key := fmt.Sprintf("/pods/%s/%s", namespace, name)

	// Get existing pod
	existingJSON, err := rs.etcdManager.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	var existingPod map[string]interface{}
	json.Unmarshal([]byte(existingJSON), &existingPod)

	// Update with new data
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Merge updates
	for k, v := range updates {
		existingPod[k] = v
	}

	// Update timestamp
	existingPod["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	updatedJSON, _ := json.Marshal(existingPod)
	err = rs.etcdManager.Put(r.Context(), key, string(updatedJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pod":     existingPod,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	key := fmt.Sprintf("/pods/%s/%s", namespace, name)
	err := rs.etcdManager.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Snapshot handlers
func (rs *RESTServer) listSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
	prefix := "/snapshots/"
	snapshots, err := rs.etcdManager.GetWithPrefix(r.Context(), prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var snapshotList []map[string]interface{}
	for key, value := range snapshots {
		var snapshot map[string]interface{}
		if err := json.Unmarshal([]byte(value), &snapshot); err != nil {
			continue
		}
		snapshotList = append(snapshotList, snapshot)
	}

	response := map[string]interface{}{
		"snapshots": snapshotList,
		"count":     len(snapshotList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) createSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name, _ := req["name"].(string)
	if name == "" {
		name = fmt.Sprintf("snapshot-%d", time.Now().Unix())
	}

	// Create snapshot
	snapshotData, err := rs.etcdManager.CreateSnapshot(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	snapshot := map[string]interface{}{
		"id":            fmt.Sprintf("snap-%d", time.Now().Unix()),
		"name":          name,
		"status":        "completed",
		"created_at":    time.Now().UTC().Format(time.RFC3339),
		"size_bytes":    len(snapshotData),
		"etcd_revision": "unknown", // Would need to get from etcd
		"checksum":      "unknown", // Would compute hash
		"metadata":      req["metadata"],
	}

	snapshotJSON, _ := json.Marshal(snapshot)
	key := fmt.Sprintf("/snapshots/%s", snapshot["id"])
	err = rs.etcdManager.Put(r.Context(), key, string(snapshotJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"snapshot": snapshot,
		"success":  true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) getSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	key := fmt.Sprintf("/snapshots/%s", id)
	snapshotJSON, err := rs.etcdManager.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Snapshot not found", http.StatusNotFound)
		return
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(snapshotJSON), &snapshot); err != nil {
		http.Error(w, "Invalid snapshot data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"snapshot": snapshot,
		"found":    true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) restoreSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// This is a simplified implementation
	// In production, you'd retrieve the snapshot data and restore it
	response := map[string]interface{}{
		"success":           false,
		"error":             "Snapshot restore not implemented",
		"restored_revision": "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) deleteSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	key := fmt.Sprintf("/snapshots/%s", id)
	err := rs.etcdManager.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Lease handlers
func (rs *RESTServer) listLeasesHandler(w http.ResponseWriter, r *http.Request) {
	prefix := "/leases/"
	leases, err := rs.etcdManager.GetWithPrefix(r.Context(), prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var leaseList []map[string]interface{}
	for key, value := range leases {
		var lease map[string]interface{}
		if err := json.Unmarshal([]byte(value), &lease); err != nil {
			continue
		}
		leaseList = append(leaseList, lease)
	}

	response := map[string]interface{}{
		"leases": leaseList,
		"count":  len(leaseList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) createLeaseHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	holder, _ := req["holder"].(string)
	ttlSeconds, _ := req["ttl_seconds"].(float64)

	if holder == "" {
		http.Error(w, "Lease holder is required", http.StatusBadRequest)
		return
	}

	if ttlSeconds <= 0 {
		ttlSeconds = 30 // Default 30 seconds
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ttlSeconds) * time.Second)

	lease := map[string]interface{}{
		"id":         fmt.Sprintf("lease-%d", time.Now().Unix()),
		"holder":     holder,
		"ttl_seconds": ttlSeconds,
		"granted_at": now.Format(time.RFC3339),
		"expires_at": expiresAt.Format(time.RFC3339),
		"metadata":   req["metadata"],
	}

	leaseJSON, _ := json.Marshal(lease)
	key := fmt.Sprintf("/leases/%s", lease["id"])
	err := rs.etcdManager.Put(r.Context(), key, string(leaseJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"lease":   lease,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) getLeaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	key := fmt.Sprintf("/leases/%s", id)
	leaseJSON, err := rs.etcdManager.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Lease not found", http.StatusNotFound)
		return
	}

	var lease map[string]interface{}
	if err := json.Unmarshal([]byte(leaseJSON), &lease); err != nil {
		http.Error(w, "Invalid lease data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"lease": lease,
		"found": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) renewLeaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	key := fmt.Sprintf("/leases/%s", id)

	// Get existing lease
	existingJSON, err := rs.etcdManager.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Lease not found", http.StatusNotFound)
		return
	}

	var lease map[string]interface{}
	json.Unmarshal([]byte(existingJSON), &lease)

	// Parse request for new TTL
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)
	newTTL, _ := req["ttl_seconds"].(float64)
	if newTTL <= 0 {
		newTTL = lease["ttl_seconds"].(float64)
	}

	// Update lease
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(newTTL) * time.Second)

	lease["ttl_seconds"] = newTTL
	lease["granted_at"] = now.Format(time.RFC3339)
	lease["expires_at"] = expiresAt.Format(time.RFC3339)

	updatedJSON, _ := json.Marshal(lease)
	err = rs.etcdManager.Put(r.Context(), key, string(updatedJSON))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"lease":   lease,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rs *RESTServer) deleteLeaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	key := fmt.Sprintf("/leases/%s", id)
	err := rs.etcdManager.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// nodeInfoHandler handles node info requests
func (rs *RESTServer) nodeInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"node_id":      "node-1", // Would get from config
		"version":     "0.1.0",
		"is_leader":   rs.etcdManager.IsLeader(),
		"leader_addr": rs.etcdManager.GetLeaderAddr(),
		"address":     rs.server.Addr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
