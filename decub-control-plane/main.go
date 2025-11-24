package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// ControlPlane represents the local control plane
type ControlPlane struct {
	etcdClient *clientv3.Client
}

// NewControlPlane creates a new control plane
func NewControlPlane(etcdEndpoints []string) (*ControlPlane, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &ControlPlane{etcdClient: cli}, nil
}

// CreateSnapshot creates an etcd snapshot
func (cp *ControlPlane) CreateSnapshot() ([]byte, error) {
	// In a real implementation, use etcd snapshot API
	// For PoC, return mock data
	mockSnapshot := map[string]interface{}{
		"version": "3.5.0",
		"data":    "mock etcd data",
		"size":    1024,
	}
	return json.Marshal(mockSnapshot)
}

// RestoreSnapshot restores an etcd snapshot
func (cp *ControlPlane) RestoreSnapshot(data []byte) error {
	// In a real implementation, restore from snapshot
	// For PoC, just log
	log.Printf("Restoring snapshot: %s", string(data))
	return nil
}

// Put puts a key-value pair
func (cp *ControlPlane) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cp.etcdClient.Put(ctx, key, value)
	return err
}

// Get gets a value by key
func (cp *ControlPlane) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cp.etcdClient.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found")
	}

	return string(resp.Kvs[0].Value), nil
}

// Watch watches for changes on a key prefix
func (cp *ControlPlane) Watch(prefix string) clientv3.WatchChan {
	return cp.etcdClient.Watch(context.Background(), prefix, clientv3.WithPrefix())
}

// Close closes the control plane
func (cp *ControlPlane) Close() error {
	return cp.etcdClient.Close()
}

// API handlers
func (cp *ControlPlane) handleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	data, err := cp.CreateSnapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (cp *ControlPlane) handleRestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := cp.RestoreSnapshot([]byte(req.Data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Snapshot restored")
}

func (cp *ControlPlane) handlePut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := cp.Put(key, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Key %s set", key)
}

func (cp *ControlPlane) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := cp.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := map[string]string{"key": key, "value": value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// Load config
	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	endpoints := viper.GetStringSlice("etcd.endpoints")

	cp, err := NewControlPlane(endpoints)
	if err != nil {
		log.Fatalf("Failed to create control plane: %v", err)
	}
	defer cp.Close()

	r := mux.NewRouter()
	r.HandleFunc("/snapshot/create", cp.handleCreateSnapshot).Methods("POST")
	r.HandleFunc("/snapshot/restore", cp.handleRestoreSnapshot).Methods("POST")
	r.HandleFunc("/kv/{key}", cp.handlePut).Methods("PUT")
	r.HandleFunc("/kv/{key}", cp.handleGet).Methods("GET")

	fmt.Println("Control plane server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
