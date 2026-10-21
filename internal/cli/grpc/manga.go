package grpc

import (
	"fmt"

	"mangahub/pkg/client"

	"github.com/spf13/cobra"
)

// mangaCmd is the manga subcommand under grpc
var mangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga operations via gRPC",
	Long:  `Query and search manga data using gRPC protocol.`,
}

// getCmd is the get subcommand under manga
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get manga details by ID",
	Long: `Query manga details from gRPC server by manga ID.

Example:
  mangahub grpc manga get --id one-piece`,
	RunE: runGetManga,
}

func runGetManga(cmd *cobra.Command, args []string) error {
	mangaID, _ := cmd.Flags().GetString("id")
	serverAddr, _ := cmd.Flags().GetString("server")

	if mangaID == "" {
		return fmt.Errorf("manga ID is required. Use --id or -i flag")
	}

	fmt.Printf("Connecting to gRPC server at %s...\n", serverAddr)

	// Create gRPC client and connect
	grpcClient := client.NewGRPCClient(serverAddr)
	if err := grpcClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer grpcClient.Close()

	fmt.Printf("Querying manga: %s\n\n", mangaID)

	// Call gRPC server
	resp, err := grpcClient.GetManga(mangaID)
	if err != nil {
		return fmt.Errorf("gRPC error: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("manga not found: %s", mangaID)
	}

	fmt.Println("✓ Successfully retrieved via gRPC")
	fmt.Println()
	fmt.Println("Manga Data:")
	fmt.Printf("  ID:       %s\n", resp.Manga.ID)
	fmt.Printf("  Title:    %s\n", resp.Manga.Title)
	fmt.Printf("  Author:   %s\n", resp.Manga.Author)
	fmt.Printf("  Status:   %s\n", resp.Manga.Status)
	fmt.Printf("  Chapters: %d\n", resp.Manga.Chapters)
	fmt.Printf("  Genres:   %s\n", client.FormatGenres(resp.Manga.Genres))

	return nil
}

// searchCmd is the search subcommand under manga
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search manga by query",
	Long: `Search manga from gRPC server using a search query.

Example:
  mangahub grpc manga search --query "naruto"`,
	RunE: runSearchManga,
}

func runSearchManga(cmd *cobra.Command, args []string) error {
	query, _ := cmd.Flags().GetString("query")
	serverAddr, _ := cmd.Flags().GetString("server")
	limit, _ := cmd.Flags().GetInt("limit")

	if query == "" {
		return fmt.Errorf("search query is required. Use --query or -q flag")
	}

	fmt.Printf("Connecting to gRPC server at %s...\n", serverAddr)

	// Create gRPC client and connect
	grpcClient := client.NewGRPCClient(serverAddr)
	if err := grpcClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer grpcClient.Close()

	fmt.Printf("Searching for: %s\n\n", query)

	// Call gRPC server
	resp, err := grpcClient.SearchManga(query, limit)
	if err != nil {
		return fmt.Errorf("gRPC error: %w", err)
	}

	fmt.Println("✓ Search completed via gRPC")
	fmt.Println()

	if len(resp.Results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("Found %d results:\n", len(resp.Results))
	for i, manga := range resp.Results {
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, manga.Title, manga.Author, manga.Status)
	}

	return nil
}

func init() {
	GRPCCmd.AddCommand(mangaCmd)
	mangaCmd.AddCommand(getCmd)
	mangaCmd.AddCommand(searchCmd)

	getCmd.Flags().StringP("id", "i", "", "Manga ID (required)")
	getCmd.Flags().StringP("server", "s", "10.238.53.72:9092", "gRPC server address")

	searchCmd.Flags().StringP("query", "q", "", "Search query (required)")
	searchCmd.Flags().StringP("server", "s", "10.238.53.72:9092", "gRPC server address")
	searchCmd.Flags().IntP("limit", "l", 10, "Maximum number of results")
}
