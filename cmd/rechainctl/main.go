package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/rechain/rechain/src/snapshot"
)

func main() {
	var rootCmd = &cobra.Command{Use: "rechainctl"}
	rootCmd.AddCommand(snapshotCmd())
	rootCmd.Execute()
}

func snapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots",
	}
	cmd.AddCommand(createCmd(), restoreCmd())
	return cmd
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create --cluster <name>",
		Short: "Create a snapshot",
	}
	cmd.Flags().String("cluster", "", "Cluster name")
	cmd.MarkFlagRequired("cluster")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cluster, _ := cmd.Flags().GetString("cluster")

		// Hardcoded for PoC
		etcdEndpoint := "http://localhost:2379"
		minioEndpoint := "localhost:9000"
		accessKey := "rechain"
		secretKey := "rechain123"
		bucket := "rechain-snapshots"

		meta, err := snapshot.CreateSnapshot(etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket, cluster)
		if err != nil {
			log.Fatalf("Failed to create snapshot: %v", err)
		}

		fmt.Printf("Snapshot created: %s\n", meta.ID)
		json.NewEncoder(os.Stdout).Encode(meta)
	}
	return cmd
}

func restoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore --id <id>",
		Short: "Restore a snapshot",
	}
	cmd.Flags().String("id", "", "Snapshot ID")
	cmd.MarkFlagRequired("id")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")

		// For simplicity, assume metadata is stored locally or fetched; in real, from catalog
		// Hardcoded for PoC
		meta := &snapshot.SnapshotMetadata{
			ID: id,
			// Other fields would be fetched from catalog
		}

		etcdEndpoint := "http://localhost:2379"
		minioEndpoint := "localhost:9000"
		accessKey := "rechain"
		secretKey := "rechain123"
		bucket := "rechain-snapshots"

		err := snapshot.RestoreSnapshot(meta, etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket)
		if err != nil {
			log.Fatalf("Failed to restore snapshot: %v", err)
		}

		fmt.Printf("Snapshot %s restored successfully\n", id)
	}
	return cmd
}
