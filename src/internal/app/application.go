package app

import (
	"log"
	"time"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/domain/entities"
)

// Application coordinates all services and follows hexagonal architecture
type Application struct {
	gameService         *GameService
	notificationService *NotificationService
	userService         *UserService
	userDataHandler     *UserDataHandler
	notificationHandler *NotificationHandler
	instanceHandler     *InstanceHandler

	// Adapters
	discordAdapter  *adapters.DiscordAdapter
	gameDataAdapter adapters.GameDataAdapter
	graphqlAdapter  adapters.GraphQLAdapter

	// Repositories
	gameRepo         entities.GameRepository
	userRepo         entities.UserRepository
	notificationRepo entities.NotificationRepository
	channelRepo      entities.ChannelRepository
	guildRepo        entities.GuildRepository

	// Channels for communication
	gameChanges chan *entities.GameChange
	commands    chan adapters.DiscordCommand

	// Control
	stopChan chan struct{}
}

// NewApplication creates a new Application instance with all dependencies
func NewApplication(
	gameRepo entities.GameRepository,
	userRepo entities.UserRepository,
	notificationRepo entities.NotificationRepository,
	channelRepo entities.ChannelRepository,
	guildRepo entities.GuildRepository,
	discordAdapter *adapters.DiscordAdapter,
	gameDataAdapter adapters.GameDataAdapter,
	graphqlAdapter adapters.GraphQLAdapter,
) *Application {

	// Create services
	gameService := NewGameService(gameRepo, gameDataAdapter)
	notificationService := NewNotificationService(notificationRepo, userRepo, channelRepo, discordAdapter)
	userService := NewUserService(userRepo)

	// Create handlers
	userDataHandler := NewUserDataHandler(userService, userRepo)
	notificationHandler := NewNotificationHandler(notificationService, discordAdapter)
	instanceHandler := NewInstanceHandler(guildRepo, channelRepo, discordAdapter)

	return &Application{
		gameService:         gameService,
		notificationService: notificationService,
		userService:         userService,
		userDataHandler:     userDataHandler,
		notificationHandler: notificationHandler,
		instanceHandler:     instanceHandler,
		discordAdapter:      discordAdapter,
		gameDataAdapter:     gameDataAdapter,
		graphqlAdapter:      graphqlAdapter,
		gameRepo:            gameRepo,
		userRepo:            userRepo,
		notificationRepo:    notificationRepo,
		channelRepo:         channelRepo,
		guildRepo:           guildRepo,
		gameChanges:         make(chan *entities.GameChange, 100),
		commands:            make(chan adapters.DiscordCommand, 100),
		stopChan:            make(chan struct{}),
	}
}

// Start starts the application
func (app *Application) Start() error {
	log.Println("Starting 18xxNotifier application...")

	// Start Discord adapter
	if err := app.discordAdapter.Start(); err != nil {
		return err
	}

	// Start background goroutines
	go app.gameUpdateLoop()
	go app.notificationLoop()
	go app.commandHandler()

	// Start handlers
	app.notificationHandler.Start()

	log.Println("Application started successfully")
	return nil
}

// Stop stops the application
func (app *Application) Stop() error {
	log.Println("Stopping 18xxNotifier application...")

	close(app.stopChan)

	// Stop handlers
	app.notificationHandler.Stop()

	if err := app.discordAdapter.Stop(); err != nil {
		return err
	}

	log.Println("Application stopped successfully")
	return nil
}

// GetGameService returns the game service for external access
func (app *Application) GetGameService() *GameService {
	return app.gameService
}

// gameUpdateLoop runs every 5 minutes to check for game updates
func (app *Application) gameUpdateLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			app.checkAllGames()
		case <-app.stopChan:
			return
		}
	}
}

// notificationLoop runs every minute to send pending notifications
func (app *Application) notificationLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := app.notificationService.SendPendingNotifications(); err != nil {
				log.Printf("Error sending notifications: %v", err)
			}
		case <-app.stopChan:
			return
		}
	}
}

// commandHandler handles Discord commands
func (app *Application) commandHandler() {
	commandChan, err := app.discordAdapter.ReceiveCommands()
	if err != nil {
		log.Printf("Error receiving commands: %v", err)
		return
	}

	for {
		select {
		case cmd := <-commandChan:
			app.handleCommand(cmd)
		case <-app.stopChan:
			return
		}
	}
}

// checkAllGames checks all tracked games for updates
func (app *Application) checkAllGames() {
	games, err := app.gameService.GetActiveGames()
	if err != nil {
		log.Printf("Error getting active games: %v", err)
		return
	}

	for _, game := range games {
		changes, err := app.gameService.UpdateGameData(game.ID)
		if err != nil {
			log.Printf("Error updating game %s: %v", game.ID, err)
			continue
		}

		// Handle all changes
		for _, change := range changes {
			if change != nil {
				// Send change to notification handler
				app.notificationHandler.HandleGameChange(change)
			}
		}
	}
}

// handleCommand processes Discord commands
func (app *Application) handleCommand(cmd adapters.DiscordCommand) {
	// First, try to handle instance management commands
	app.instanceHandler.HandleDiscordCommand(cmd)

	// Then handle other commands
	switch cmd.Command {
	case "track":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.gameService.TrackGame(gameID); err != nil {
				log.Printf("Error tracking game %s: %v", gameID, err)
			} else {
				log.Printf("Successfully started tracking game %s", gameID)
			}
		} else {
			log.Printf("Track command requires game ID")
		}
	case "link":
		if len(cmd.Args) > 0 {
			account := cmd.Args[0]
			if err := app.userDataHandler.LinkUserToAccount(cmd.UserID, account); err != nil {
				log.Printf("Error linking account: %v", err)
			} else {
				log.Printf("Successfully linked user %s to account %s", cmd.UserID, account)
			}
		} else {
			log.Printf("Link command requires 18xx account name")
		}
	case "subscribe":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.userDataHandler.SubscribeUserToGame(cmd.UserID, gameID); err != nil {
				log.Printf("Error subscribing to game: %v", err)
			} else {
				log.Printf("Successfully subscribed user %s to game %s", cmd.UserID, gameID)
			}
		} else {
			log.Printf("Subscribe command requires game ID")
		}
	case "unsubscribe":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.userDataHandler.UnsubscribeUserFromGame(cmd.UserID, gameID); err != nil {
				log.Printf("Error unsubscribing from game: %v", err)
			} else {
				log.Printf("Successfully unsubscribed user %s from game %s", cmd.UserID, gameID)
			}
		} else {
			log.Printf("Unsubscribe command requires game ID")
		}
	case "games":
		games, err := app.gameService.GetActiveGames()
		if err != nil {
			log.Printf("Error getting active games: %v", err)
		} else {
			log.Printf("Active games: %d", len(games))
			for _, game := range games {
				log.Printf("  - %s (active: %s, finished: %v)", game.ID, game.ActivePlayer, game.Finished)
			}
		}
	default:
		log.Printf("Unknown command: %s", cmd.Command)
	}
}
