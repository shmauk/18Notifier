package app

import (
	"fmt"
	"log"
	"strings"
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
	case "help":
		helpMessage := "**18xxNotifier Commands:**\n\n" +
			"**User Registration:**\n" +
			"• `!18xx register` - Register your Discord account\n" +
			"• `!18xx link <username>` - Link your 18xx.games username\n" +
			"• `!18xx unlink <username>` - Unlink an 18xx account\n\n" +
			"**Game Subscriptions:**\n" +
			"• `!18xx subscribe <game_id>` - Subscribe to game notifications\n" +
			"• `!18xx unsubscribe <game_id>` - Unsubscribe from game notifications\n" +
			"• `!18xx subscriptions` - List your subscriptions\n\n" +
			"**Game Tracking (Admin Only):**\n" +
			"• `!18xx track <game_id>` - Start tracking a game in this channel\n" +
			"• `!18xx untrack <game_id>` - Stop tracking a game\n" +
			"• `!18xx tracked` - List tracked games\n\n" +
			"**Utility:**\n" +
			"• `!18xx help` - Show this help message\n" +
			"• `!18xx test` - Send a test notification\n" +
			"• `!18xx status` - Show bot status"

		// Send help message to the channel
		if err := app.discordAdapter.SendNotification(cmd.ChannelID, helpMessage, nil); err != nil {
			log.Printf("Error sending help message: %v", err)
		}

	case "test":
		testMessage := "✅ **18xxNotifier is working!** This is a test notification."
		if err := app.discordAdapter.SendNotification(cmd.ChannelID, testMessage, nil); err != nil {
			log.Printf("Error sending test message: %v", err)
		}

	case "track":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.gameService.TrackGame(gameID, cmd.ChannelID, cmd.GuildID); err != nil {
				log.Printf("Error tracking game %s: %v", gameID, err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to track game: "+err.Error(), nil)
			} else {
				log.Printf("Successfully started tracking game %s", gameID)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully started tracking game: **"+gameID+"**", nil)
			}
		} else {
			log.Printf("Track command requires game ID")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide a game ID. Usage: `!18xx track <game_id>`", nil)
		}
	case "untrack":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.gameService.StopTrackingGame(gameID); err != nil {
				log.Printf("Error untracking game %s: %v", gameID, err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to untrack game: "+err.Error(), nil)
			} else {
				log.Printf("Successfully stopped tracking game %s", gameID)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully stopped tracking game: **"+gameID+"**", nil)
			}
		} else {
			log.Printf("Untrack command requires game ID")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide a game ID. Usage: `!18xx untrack <game_id>`", nil)
		}
	case "link":
		if len(cmd.Args) > 0 {
			account := cmd.Args[0]
			if err := app.userDataHandler.LinkUserToAccount(cmd.UserID, account); err != nil {
				log.Printf("Error linking account: %v", err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to link account: "+err.Error(), nil)
			} else {
				log.Printf("Successfully linked user %s to account %s", cmd.UserID, account)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully linked your Discord account to 18xx.games account: **"+account+"**", nil)
			}
		} else {
			log.Printf("Link command requires 18xx account name")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide your 18xx.games username. Usage: `!18xx link <username>`", nil)
		}
	case "subscribe":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]

			// First, ensure the game exists in the database by fetching it from the API
			game, err := app.gameDataAdapter.FetchGameData(gameID)
			if err != nil {
				log.Printf("Error fetching game data for subscription: %v", err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to subscribe: Game not found or API error: "+err.Error(), nil)
				break
			}

			// Check if game already exists in database, if not save it
			existingGame, err := app.gameRepo.GetGame(gameID)
			if err != nil || existingGame == nil {
				log.Printf("Game %s not in database, creating it for subscription", gameID)
				// Save the game with the current channel as the tracking channel
				if err := app.gameRepo.SaveGame(game, cmd.ChannelID, cmd.GuildID); err != nil {
					log.Printf("Error saving game for subscription: %v", err)
					_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to subscribe: Could not save game to database: "+err.Error(), nil)
					break
				}
			}

			// Now subscribe the user to the game
			if err := app.userDataHandler.SubscribeUserToGame(cmd.UserID, gameID); err != nil {
				log.Printf("Error subscribing to game: %v", err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to subscribe to game: "+err.Error(), nil)
			} else {
				log.Printf("Successfully subscribed user %s to game %s", cmd.UserID, gameID)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully subscribed to game: **"+gameID+"**", nil)
			}
		} else {
			log.Printf("Subscribe command requires game ID")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide a game ID. Usage: `!18xx subscribe <game_id>`", nil)
		}
	case "unsubscribe":
		if len(cmd.Args) > 0 {
			gameID := cmd.Args[0]
			if err := app.userDataHandler.UnsubscribeUserFromGame(cmd.UserID, gameID); err != nil {
				log.Printf("Error unsubscribing from game: %v", err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to unsubscribe from game: "+err.Error(), nil)
			} else {
				log.Printf("Successfully unsubscribed user %s from game %s", cmd.UserID, gameID)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully unsubscribed from game: **"+gameID+"**", nil)
			}
		} else {
			log.Printf("Unsubscribe command requires game ID")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide a game ID. Usage: `!18xx unsubscribe <game_id>`", nil)
		}
	case "games":
		games, err := app.gameService.GetActiveGames()
		if err != nil {
			log.Printf("Error getting active games: %v", err)
		} else {
			log.Printf("Active games: %d", len(games))
			for _, game := range games {
				log.Printf("  - %s (active: %v, finished: %v)", game.ID, game.ActivePlayers, game.Finished)
			}
		}
	case "register":
		err := app.userDataHandler.RegisterUser(cmd.UserID)
		if err != nil {
			if strings.Contains(err.Error(), "user already exists") {
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "ℹ️ You are already registered.", nil)
			} else {
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Registration failed: "+err.Error(), nil)
			}
		} else {
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ You are now registered!", nil)
		}
	case "unlink":
		if len(cmd.Args) > 0 {
			account := cmd.Args[0]
			if err := app.userDataHandler.UnlinkUserFromAccount(cmd.UserID, account); err != nil {
				log.Printf("Error unlinking account: %v", err)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to unlink account: "+err.Error(), nil)
			} else {
				log.Printf("Successfully unlinked user %s from account %s", cmd.UserID, account)
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "✅ Successfully unlinked your Discord account from 18xx.games account: **"+account+"**", nil)
			}
		} else {
			log.Printf("Unlink command requires 18xx account name")
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Please provide your 18xx.games username. Usage: `!18xx unlink <username>`", nil)
		}
	case "subscriptions":
		subscriptions, err := app.userDataHandler.GetUserSubscriptions(cmd.UserID)
		if err != nil {
			log.Printf("Error getting user subscriptions: %v", err)
			_ = app.discordAdapter.SendNotification(cmd.ChannelID, "❌ Failed to get subscriptions: "+err.Error(), nil)
		} else {
			if len(subscriptions) == 0 {
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, "📋 You are not subscribed to any games.", nil)
			} else {
				message := "📋 **Your Game Subscriptions:**\n\n"
				for _, gameID := range subscriptions {
					message += fmt.Sprintf("• Game **%s**\n", gameID)
				}
				_ = app.discordAdapter.SendNotification(cmd.ChannelID, message, nil)
			}
		}
	default:
		log.Printf("Unknown command: %s", cmd.Command)
	}
}
