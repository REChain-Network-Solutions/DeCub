package main

import (
	"os"

	"github.com/rechain/rechain/src/gcl"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mockGCL := gcl.NewMockGCL(port)
	gcl.GlobalMockGCL = mockGCL
	mockGCL.StartServer()
}
