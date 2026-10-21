package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "mangahub/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec implements grpc encoding.Codec using JSON
type JSONCodec struct{}

func (JSONCodec) Name() string { return "json" }

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GRPCClient represents a gRPC client for manga service
type GRPCClient struct {
	ServerAddr string
	conn       *grpc.ClientConn
	client     pb.MangaServiceClient
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(serverAddr string) *GRPCClient {
	return &GRPCClient{
		ServerAddr: serverAddr,
	}
}

// Connect connects to the gRPC server
func (c *GRPCClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.ServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype("json")),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	c.conn = conn
	c.client = pb.NewMangaServiceClient(conn)
	return nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GRPCMangaResponse wraps manga response with success flag
type GRPCMangaResponse struct {
	Success bool
	Manga   *pb.MangaResponse
}

// GetManga retrieves manga by ID
func (c *GRPCClient) GetManga(mangaID string) (*GRPCMangaResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.GetManga(ctx, &pb.MangaRequest{ID: mangaID})
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	return &GRPCMangaResponse{
		Success: resp.ID != "",
		Manga:   resp,
	}, nil
}

// SearchManga searches for manga
func (c *GRPCClient) SearchManga(query string, limit int) (*pb.SearchResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.SearchManga(ctx, &pb.SearchRequest{
		Title: query,
		Limit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search manga: %w", err)
	}

	return resp, nil
}

// UpdateProgress updates reading progress
func (c *GRPCClient) UpdateProgress(userID, mangaID string, chapter int) (*pb.UpdateProgressResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.UpdateProgress(ctx, &pb.UpdateProgressRequest{
		UserID:  userID,
		MangaID: mangaID,
		Chapter: int32(chapter),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update progress: %w", err)
	}

	return resp, nil
}

// GetTop10Manga retrieves top 10 manga
func (c *GRPCClient) GetTop10Manga() (*pb.Top10Response, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.GetTop10Manga(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get top manga: %w", err)
	}

	return resp, nil
}

// FormatGenres formats genres as a string
func FormatGenres(genres []string) string {
	if len(genres) == 0 {
		return "N/A"
	}
	return strings.Join(genres, ", ")
}
