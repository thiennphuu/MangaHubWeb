package notify

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to notifications",
	Long: `Subscribe to manga chapter release notifications via UDP server.

This command connects to the UDP notification server and registers
the client to receive real-time chapter release notifications.

Examples:
  mangahub notify subscribe
  mangahub notify subscribe --manga-id one-piece
  mangahub notify subscribe --all
  mangahub notify subscribe --listen`,
	RunE: runSubscribe,
}

func runSubscribe(cmd *cobra.Command, args []string) error {
	logger := utils.NewLogger()
	mangaID, _ := cmd.Flags().GetString("manga-id")
	all, _ := cmd.Flags().GetBool("all")
	listen, _ := cmd.Flags().GetBool("listen")
	serverAddr, _ := cmd.Flags().GetString("server")

	// Check if user is logged in
	session, err := loadSession()
	if err != nil {
		logger.Warn("user is not logged in")
		fmt.Println("You are not logged in.")
		fmt.Println("\nPlease login first:")
		fmt.Println("  mangahub auth login --username <username>")
		return nil
	}

	// Get notification service for database operations
	notifySvc, err := getNotificationService()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Connect to UDP server
	fmt.Printf("Connecting to UDP notification server at %s...\n", serverAddr)
	udpClient := client.NewUDPClient(serverAddr)
	err = udpClient.Connect()
	if err != nil {
		logger.Warn("could not connect to UDP server: %v", err)
		fmt.Println("  Continuing with local database subscription only.")
		fmt.Println("  Make sure UDP server is running: go run ./cmd/udp-server")
	} else {
		defer udpClient.Close()

		// Register with UDP server
		err = udpClient.Register()
		if err != nil {
			logger.Warn("failed to register with UDP server: %v", err)
		} else {
			fmt.Println("✓ Registered with UDP notification server")
			fmt.Printf("  Local address: %s\n", udpClient.GetLocalAddr())
		}
	}

	// Handle subscription logic
	if mangaID != "" {
		// Subscribe to specific manga
		err = notifySvc.Subscribe(session.UserID, mangaID)
		if err != nil {
			return fmt.Errorf("failed to subscribe: %w", err)
		}
		fmt.Printf("✓ Subscribed to notifications for '%s'\n", mangaID)
	} else if all {
		// Subscribe to all manga in library
		libSvc, err := getLibraryService()
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}

		library, err := libSvc.GetLibrary(session.UserID, 100, 0)
		if err != nil {
			return fmt.Errorf("failed to get library: %w", err)
		}

		count := 0
		for _, entry := range library {
			if entry.Status == "reading" || entry.Status == "plan-to-read" {
				err = notifySvc.Subscribe(session.UserID, entry.MangaID)
				if err == nil {
					count++
				}
			}
		}
		fmt.Printf("✓ Subscribed to notifications for %d manga\n", count)
	} else {
		// Enable general notifications
		err = notifySvc.EnableNotifications(session.UserID)
		if err != nil {
			return fmt.Errorf("failed to enable notifications: %w", err)
		}
		fmt.Println("✓ Subscribed to chapter release notifications")
	}

	// Show notification preferences
	fmt.Println("\nNotification Preferences:")
	prefs, _ := notifySvc.GetPreferences(session.UserID)
	if prefs != nil {
		fmt.Printf("  Chapter releases: %s\n", boolToEnabled(prefs.ChapterReleases))
		fmt.Printf("  Email notifications: %s\n", boolToEnabled(prefs.EmailNotifications))
		fmt.Printf("  Sound enabled: %s\n", boolToEnabled(prefs.SoundEnabled))
	} else {
		fmt.Println("  Chapter releases: enabled")
		fmt.Println("  Email notifications: enabled")
	}

	// If listen mode is enabled, keep listening for notifications
	if listen && udpClient != nil {
		fmt.Println("\n─────────────────────────────────────────────────────────────")
		fmt.Println("77+C to stop)")
		fmt.Println("─────────────────────────────────────────────────────────────")

		// Set up signal handler for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\n\nStopping notification listener...")
			udpClient.Close()
			os.Exit(0)
		}()

		// Listen for notifications
		err = udpClient.Listen(func(payload models.NotificationPayload) {
			// Skip system messages (registration confirmations)
			if payload.Type == "registered" || payload.Type == "unregistered" {
				return // Don't display system confirmations
			}

			timestamp := time.Unix(payload.Timestamp, 0).Format("15:04:05")
			fmt.Printf("\n[%s] 🔔 %s\n", timestamp, payload.Type)
			if payload.MangaID != "" {
				fmt.Printf("  Manga: %s\n", payload.MangaID)
			}
			fmt.Printf("  Message: %s\n", payload.Message)
		})
		if err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

func boolToEnabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func init() {
	NotifyCmd.AddCommand(subscribeCmd)
	subscribeCmd.Flags().StringP("manga-id", "m", "", "Subscribe to specific manga")
	subscribeCmd.Flags().BoolP("all", "a", false, "Subscribe to all manga in library")
	subscribeCmd.Flags().BoolP("listen", "l", false, "Keep listening for notifications after subscribing")
	subscribeCmd.Flags().StringP("server", "s", "10.238.53.72:9091", "UDP server address")
}
