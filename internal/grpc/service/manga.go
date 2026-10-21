package service

import (
	"context"
	"fmt"

	"mangahub/internal/manga"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
	pb "mangahub/proto"
)

// MangaService implements the gRPC MangaService
type MangaService struct {
	mangaService *manga.Service
	logger       *utils.Logger
}

// NewMangaService creates a new gRPC manga service
func NewMangaService(db *database.Database, logger *utils.Logger) *MangaService {
	return &MangaService{
		mangaService: manga.NewService(db),
		logger:       logger,
	}
}

// GetManga retrieves a manga by ID
func (s *MangaService) GetManga(ctx context.Context, req *pb.MangaRequest) (*pb.MangaResponse, error) {
	m, err := s.mangaService.GetByID(req.ID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to get manga: %v", err))
		return nil, fmt.Errorf("manga not found")
	}

	return &pb.MangaResponse{
		ID:       m.ID,
		Title:    m.Title,
		Author:   m.Author,
		Status:   m.Status,
		Chapters: int32(m.TotalChapters),
		Synopsis: m.Description,
		Genres:   m.Genres,
	}, nil
}

// SearchManga searches for manga
func (s *MangaService) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	filter := &models.MangaFilter{
		Query:  req.Title,
		Author: req.Author,
		Status: req.Status,
		Genres: req.Genres,
		Limit:  int(req.Limit),
	}

	if filter.Limit == 0 {
		filter.Limit = 10
	}

	results, err := s.mangaService.Search(filter)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to search manga: %v", err))
		return nil, fmt.Errorf("search failed")
	}

	var mangaResults []*pb.MangaResponse
	for _, m := range results.Manga {
		mangaResults = append(mangaResults, &pb.MangaResponse{
			ID:       m.ID,
			Title:    m.Title,
			Author:   m.Author,
			Status:   m.Status,
			Chapters: int32(m.TotalChapters),
			Synopsis: m.Description,
			Genres:   m.Genres,
		})
	}

	return &pb.SearchResponse{
		Results: mangaResults,
	}, nil
}

// UpdateProgress updates reading progress
func (s *MangaService) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	// TODO: Implement progress update in database
	s.logger.Info(fmt.Sprintf("Progress updated for user %s on manga %s", req.UserID, req.MangaID))

	return &pb.UpdateProgressResponse{
		Success: true,
		Message: "Progress updated successfully",
	}, nil
}

// GetTop10Manga retrieves top 10 manga
func (s *MangaService) GetTop10Manga(ctx context.Context, req *pb.Empty) (*pb.Top10Response, error) {
	mangaList, err := s.mangaService.List(10, 0)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to get top manga: %v", err))
		return nil, fmt.Errorf("failed to get manga")
	}

	var mangaResults []*pb.MangaResponse
	for _, m := range mangaList {
		mangaResults = append(mangaResults, &pb.MangaResponse{
			ID:       m.ID,
			Title:    m.Title,
			Author:   m.Author,
			Status:   m.Status,
			Chapters: int32(m.TotalChapters),
			Synopsis: m.Description,
			Genres:   m.Genres,
		})
	}

	return &pb.Top10Response{
		Rankings: mangaResults,
	}, nil
}
