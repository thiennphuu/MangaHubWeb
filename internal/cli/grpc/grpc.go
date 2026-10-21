package grpc

import "github.com/spf13/cobra"

// GRPCCmd is the main gRPC command
var GRPCCmd = &cobra.Command{
	Use:   "grpc",
	Short: "gRPC service operations",
	Long:  `Query and manipulate manga data via gRPC service calls.`,
}
