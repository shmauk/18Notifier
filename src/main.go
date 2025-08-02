package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/app"
	"github.com/18xxnotifier/internal/domain/entities"
)

// Mock implementations for development (keeping some for now)
type MockUserRepository struct{}
type MockNotificationRepository struct{}
type MockChannelRepository struct{}
type MockGuildRepository struct{}
type MockDiscordAdapter struct{}

// Mock implementations
func (m *MockUserRepository) GetUser(discordID string) (*entities.User, error)       { return nil, nil }
func (m *MockUserRepository) SaveUser(user *entities.User) error                     { return nil }
func (m *MockUserRepository) UpdateUser(user *entities.User) error                   { return nil }
func (m *MockUserRepository) DeleteUser(discordID string) error                      { return nil }
func (m *MockUserRepository) GetAllUsers() ([]*entities.User, error)                 { return nil, nil }
func (m *MockUserRepository) GetUsersByGame(gameID string) ([]*entities.User, error) { return nil, nil }

func (m *MockNotificationRepository) GetNotification(id string) (*entities.Notification, error) {
	return nil, nil
}
func (m *MockNotificationRepository) SaveNotification(notification *entities.Notification) error {
	return nil
}
func (m *MockNotificationRepository) UpdateNotification(notification *entities.Notification) error {
	return nil
}
func (m *MockNotificationRepository) DeleteNotification(id string) error { return nil }
func (m *MockNotificationRepository) GetUnsentNotifications() ([]*entities.Notification, error) {
	return nil, nil
}
func (m *MockNotificationRepository) GetNotificationsByGame(gameID string) ([]*entities.Notification, error) {
	return nil, nil
}

func (m *MockChannelRepository) GetChannel(id string) (*entities.Channel, error) { return nil, nil }
func (m *MockChannelRepository) SaveChannel(channel *entities.Channel) error     { return nil }
func (m *MockChannelRepository) UpdateChannel(channel *entities.Channel) error   { return nil }
func (m *MockChannelRepository) DeleteChannel(id string) error                   { return nil }
func (m *MockChannelRepository) GetAllChannels() ([]*entities.Channel, error)    { return nil, nil }
func (m *MockChannelRepository) GetChannelsByGuild(guildID string) ([]*entities.Channel, error) {
	return nil, nil
}
func (m *MockChannelRepository) GetChannelsByGame(gameID string) ([]*entities.Channel, error) {
	return nil, nil
}

func (m *MockGuildRepository) GetGuild(id string) (*entities.Guild, error) { return nil, nil }
func (m *MockGuildRepository) SaveGuild(guild *entities.Guild) error       { return nil }
func (m *MockGuildRepository) UpdateGuild(guild *entities.Guild) error     { return nil }
func (m *MockGuildRepository) DeleteGuild(id string) error                 { return nil }
func (m *MockGuildRepository) GetAllGuilds() ([]*entities.Guild, error)    { return nil, nil }

func (m *MockDiscordAdapter) SendNotification(channelID string, message string, userMentions []string) error {
	return nil
}
func (m *MockDiscordAdapter) ReceiveCommands() (<-chan adapters.DiscordCommand, error) {
	return nil, nil
}
func (m *MockDiscordAdapter) Start() error { return nil }
func (m *MockDiscordAdapter) Stop() error  { return nil }

func main() {
	// Get environment variables
	dgraphEndpoint := os.Getenv("DGRAPH_ENDPOINT")
	if dgraphEndpoint == "" {
		dgraphEndpoint = "http://dgraph:8080"
	}

	// Create real adapters
	eighteenxxAdapter := adapters.NewEighteenxxAPIAdapter()
	graphqlAdapter := adapters.NewDGraphGraphQLAdapter(dgraphEndpoint)

	// Create repositories
	gameRepo := adapters.NewGraphQLGameRepository(graphqlAdapter)
	userRepo := adapters.NewGraphQLUserRepository(graphqlAdapter)
	notificationRepo := adapters.NewGraphQLNotificationRepository(graphqlAdapter)
	channelRepo := adapters.NewGraphQLChannelRepository(graphqlAdapter)
	guildRepo := adapters.NewGraphQLGuildRepository(graphqlAdapter)

	// Create real Discord adapter
	discordAdapter := adapters.NewDiscordAdapter()

	// Create application with hexagonal architecture
	application := app.NewApplication(
		gameRepo,
		userRepo,
		notificationRepo,
		channelRepo,
		guildRepo,
		discordAdapter,
		eighteenxxAdapter,
		graphqlAdapter,
	)

	// Create game data handler
	gameDataHandler := app.NewGameDataHandler(application.GetGameService(), 5*time.Minute)

	// Start the application
	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Start the game data handler
	go gameDataHandler.Start()

	// Set up HTTP server for health checks
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("18xxNotifier Service is running"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Starting 18xxNotifier service on port %s\n", port)
	fmt.Printf("DGraph endpoint: %s\n", dgraphEndpoint)

	// Start HTTP server in a goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	gameDataHandler.Stop()
	if err := application.Stop(); err != nil {
		log.Printf("Error stopping application: %v", err)
	}
}
