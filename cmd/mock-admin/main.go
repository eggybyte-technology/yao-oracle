package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/eggybyte-technology/yao-oracle/core/utils"
	"github.com/eggybyte-technology/yao-oracle/internal/dashboard"
)

// main is the entry point for the mock-admin service.
//
// The mock-admin service provides a test dashboard backend with:
//   - gRPC streaming API for real-time metrics
//   - Mock data generation for testing UI
//   - No dependencies on real Kubernetes cluster
//
// Usage:
//
//	mock-admin --grpc-port=9090 --password=admin123 --refresh-interval=5
func main() {
	// Parse command line flags
	grpcPort := flag.Int("grpc-port", 9090, "gRPC server port")
	password := flag.String("password", "admin123", "Dashboard password")
	refreshInterval := flag.Int("refresh-interval", 5, "Metrics refresh interval in seconds")
	flag.Parse()

	logger := utils.NewLogger("mock-admin")

	// Print banner
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         🎯 Yao-Oracle Mock Admin Service (Test Mode)       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	logger.Info("Starting mock-admin service...")
	logger.Info("Configuration:")
	logger.Info("  - gRPC Port: %d", *grpcPort)
	logger.Info("  - Refresh Interval: %d seconds", *refreshInterval)
	logger.Info("  - Dashboard Password: %s", *password)
	logger.Info("  - Test Mode: Enabled (Mock Data)")
	logger.Info("")

	// Create mock configuration informer
	mockInformer := dashboard.NewMockConfigInformer(*password)

	// Create gRPC dashboard server in test mode
	dashboardServer := dashboard.NewDashboardGRPCServer(mockInformer, *refreshInterval, true)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	dashboard.RegisterDashboardServer(grpcServer, dashboardServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	// Start gRPC server in goroutine
	go func() {
		logger.Info("✅ gRPC server listening on localhost:%d", *grpcPort)
		logger.Info("📡 Dashboard clients can now connect and stream metrics")
		logger.Info("🔄 Mock data refreshing every %d seconds", *refreshInterval)
		logger.Info("")
		logger.Info("Ready to accept connections...")
		logger.Info("─────────────────────────────────────────────────────────────")

		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("")
	logger.Info("─────────────────────────────────────────────────────────────")
	logger.Info("Shutting down mock-admin service...")

	// Graceful shutdown
	grpcServer.GracefulStop()
	dashboardServer.Stop()

	logger.Info("✅ Mock-admin service stopped gracefully")
}
