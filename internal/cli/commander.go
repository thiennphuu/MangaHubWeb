package cli

import (
	"fmt"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
)

// Commander holds CLI command handlers
type Commander struct {
	httpClient *client.HTTPClient
	logger     *utils.Logger
	prompt     *utils.Prompt
}

// NewCommander creates a new CLI commander
func NewCommander(logger *utils.Logger) *Commander {
	return &Commander{
		httpClient: client.NewHTTPClient("http://localhost:8080", ""),
		logger:     logger,
		prompt:     utils.NewPrompt(),
	}
}

// Register handles user registration command
func (cmd *Commander) Register() error {
	cmd.logger.Info("=== User Registration ===")

	username, _ := cmd.prompt.String("Username: ")
	if err := utils.ValidateUsername(username); err != nil {
		return err
	}

	email, _ := cmd.prompt.String("Email: ")
	if err := utils.ValidateEmail(email); err != nil {
		return err
	}

	password, _ := cmd.prompt.Password("Password: ")
	if err := utils.ValidatePassword(password); err != nil {
		return err
	}

	confirmPassword, _ := cmd.prompt.Password("Confirm Password: ")
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Use HTTP client to register
	if _, err := cmd.httpClient.Register(username, email, password); err != nil {
		return err
	}

	cmd.logger.Info("User registered successfully!")
	return nil
}

// Login handles user login command
func (cmd *Commander) Login() error {
	cmd.logger.Info("=== User Login ===")

	username, _ := cmd.prompt.String("Username: ")
	password, _ := cmd.prompt.Password("Password: ")

	resp, err := cmd.httpClient.Login(username, password)
	if err != nil {
		return err
	}

	cmd.httpClient.SetToken(resp.Token)
	cmd.logger.Info(fmt.Sprintf("Login successful! Token expires at: %s", resp.ExpiresAt))
	return nil
}

// SearchManga handles manga search command
func (cmd *Commander) SearchManga() error {
	cmd.logger.Info("=== Manga Search ===")

	title, _ := cmd.prompt.String("Title (or part of): ")

	filter := &models.MangaFilter{Query: title}
	results, err := cmd.httpClient.SearchManga(filter)
	if err != nil {
		return err
	}

	if len(results.Manga) == 0 {
		cmd.logger.Info("No manga found")
		return nil
	}

	cmd.logger.Info(fmt.Sprintf("Found %d manga:", len(results.Manga)))
	for i, m := range results.Manga {
		fmt.Printf("%d. %s by %s (Status: %s)\n", i+1, m.Title, m.Author, m.Status)
	}

	return nil
}

// ViewManga handles viewing manga details
func (cmd *Commander) ViewManga() error {
	cmd.logger.Info("=== View Manga ===")

	mangaID, _ := cmd.prompt.String("Manga ID: ")

	m, err := cmd.httpClient.GetManga(mangaID)
	if err != nil {
		return err
	}

	fmt.Printf("Title: %s\n", m.Title)
	fmt.Printf("Author: %s\n", m.Author)
	fmt.Printf("Status: %s\n", m.Status)
	fmt.Printf("Chapters: %d\n", m.TotalChapters)
	fmt.Printf("Description: %s\n", m.Description)
	fmt.Printf("Genres: %v\n", m.Genres)

	return nil
}

// AddToLibrary handles adding manga to library
func (cmd *Commander) AddToLibrary() error {
	cmd.logger.Info("=== Add to Library ===")

	mangaID, _ := cmd.prompt.String("Manga ID: ")
	status, _ := cmd.prompt.String("Status (reading/completed/on-hold/dropped): ")

	if err := cmd.httpClient.AddToLibrary(mangaID, status, 0, ""); err != nil {
		return err
	}

	cmd.logger.Info("Manga added to library")
	return nil
}

// ViewLibrary handles viewing user's library
func (cmd *Commander) ViewLibrary() error {
	cmd.logger.Info("=== Your Library ===")

	status, _ := cmd.prompt.String("Filter by status (or leave blank): ")

	lib, err := cmd.httpClient.GetLibrary(status, 50, 0)
	if err != nil {
		return err
	}

	if len(lib) == 0 {
		cmd.logger.Info("Your library is empty")
		return nil
	}

	cmd.logger.Info(fmt.Sprintf("Library entries: %d", len(lib)))
	for i, p := range lib {
		fmt.Printf("%d. Manga %s - Chapter %d (%s)\n", i+1, p.MangaID, p.CurrentChapter, p.Status)
	}

	return nil
}

// UpdateProgress handles updating reading progress
func (cmd *Commander) UpdateProgress() error {
	cmd.logger.Info("=== Update Progress ===")

	mangaID, _ := cmd.prompt.String("Manga ID: ")
	chapterStr, _ := cmd.prompt.String("Current Chapter: ")
	statusStr, _ := cmd.prompt.String("Status: ")
	ratingStr, _ := cmd.prompt.String("Rating (1-5): ")

	chapter := 0
	fmt.Sscanf(chapterStr, "%d", &chapter)

	rating := 0
	fmt.Sscanf(ratingStr, "%d", &rating)

	if err := cmd.httpClient.UpdateProgress(mangaID, chapter, statusStr, rating, ""); err != nil {
		return err
	}

	cmd.logger.Info("Progress updated successfully")
	return nil
}
	
// RemoveFromLibrary handles removing manga from library
func (cmd *Commander) RemoveFromLibrary() error {
	cmd.logger.Info("=== Remove from Library ===")

	mangaID, _ := cmd.prompt.String("Manga ID: ")

	if err := cmd.httpClient.RemoveFromLibrary(mangaID); err != nil {
		return err
	}

	cmd.logger.Info("Manga removed from library")
	return nil
}

// JoinChat handles joining a chat room
func (cmd *Commander) JoinChat() error {
	cmd.logger.Info("=== Join Chat ===")

	room, _ := cmd.prompt.String("Room name: ")
	cmd.logger.Info(fmt.Sprintf("Joining room: %s", room))

	// TODO: Implement WebSocket connection
	return nil
}

// SendMessage handles sending a chat message
func (cmd *Commander) SendMessage() error {
	cmd.logger.Info("=== Send Message ===")

	message, _ := cmd.prompt.String("Message: ")

	// TODO: Implement sending message via WebSocket
	fmt.Printf("Message sent: %s\n", message)
	return nil
}

// ViewStats handles viewing reading statistics
func (cmd *Commander) ViewStats() error {
	cmd.logger.Info("=== Reading Statistics ===")

	// TODO: Implement statistics calculation and display
	cmd.logger.Info("Statistics feature coming soon")
	return nil
}

// ExportLibrary handles exporting library data
func (cmd *Commander) ExportLibrary() error {
	cmd.logger.Info("=== Export Library ===")

	format, _ := cmd.prompt.String("Format (json/csv): ")

	// TODO: Implement export functionality
	fmt.Printf("Exporting library as %s...\n", format)
	return nil
}

// SyncProgress handles progress synchronization
func (cmd *Commander) SyncProgress() error {
	cmd.logger.Info("=== Sync Progress ===")

	// TODO: Implement TCP-based progress synchronization
	cmd.logger.Info("Connecting to sync server...")
	return nil
}

// ShowNotifications handles showing notifications
func (cmd *Commander) ShowNotifications() error {
	cmd.logger.Info("=== Notifications ===")

	// TODO: Implement notification retrieval via UDP
	cmd.logger.Info("No new notifications")
	return nil
}

// ManageServer handles server management commands
func (cmd *Commander) ManageServer(action string) error {
	cmd.logger.Info(fmt.Sprintf("=== Server Management: %s ===", action))

	// TODO: Implement server control commands (start/stop/restart/status)
	return nil
}

// ShowHelp displays help information
func (cmd *Commander) ShowHelp() {
	help := `
MangaHub - Manga Tracking System
================================

Commands:
  register              - Create a new account
  login                 - Login to your account
  search                - Search for manga
  view                  - View manga details
  add                   - Add manga to library
  library               - View your library
  update                - Update reading progress
  remove                - Remove manga from library
  stats                 - View reading statistics
  export                - Export library data
  sync                  - Synchronize progress
  notifications         - Check notifications
  chat                  - Join chat room
  send                  - Send chat message
  help                  - Show this help message
  exit                  - Exit the application

For more information, visit: https://github.com/yourusername/mangahub
`
	fmt.Println(help)
}

// ExecuteCommand executes a command
func (cmd *Commander) ExecuteCommand(command string) error {
	switch command {
	case "register":
		return cmd.Register()
	case "login":
		return cmd.Login()
	case "search":
		return cmd.SearchManga()
	case "view":
		return cmd.ViewManga()
	case "add":
		return cmd.AddToLibrary()
	case "library":
		return cmd.ViewLibrary()
	case "update":
		return cmd.UpdateProgress()
	case "remove":
		return cmd.RemoveFromLibrary()
	case "stats":
		return cmd.ViewStats()
	case "export":
		return cmd.ExportLibrary()
	case "sync":
		return cmd.SyncProgress()
	case "notifications":
		return cmd.ShowNotifications()
	case "chat":
		return cmd.JoinChat()
	case "send":
		return cmd.SendMessage()
	case "help":
		cmd.ShowHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
