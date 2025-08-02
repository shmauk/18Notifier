# Tasks

Tasks are sorted into sections by type and labelled accordingly.

## Adapters

The service will interface with external systems using adapters as outlined in architecture.mermaid.

### ADP-001 GraphQL Adapter
Description: Create an extendable adapter for graphql that can perform queries and mutations on the dgraph container
Acceptance Criteria:
    - Strictly uses GraphQL queries and mutations to speak to the dgraph database

### ADP-002 Discord Adapter
Description: Create an adapter for a discord bot that can send notifications and receive commands from a discord channel in a discord server
Acceptance Criteria:
    - Sends notifications to a channel with a user mention
    - Receives commands from discord users in a discord channel

### ADP-003 18xx API Adapter
Description: Create an adapter that calls the api at https://18xx.games/api/game/<gameId> to receive the game data
Acceptance Criteria:
    - Can request game data from the api
    - Provides any error response to the app layer

## Application

Sends and receives data to and from the adapter layer and handles the data

### APP-001
Description: Set up an app and domain layer structure to follow hexagonal architecture best practices
Acceptance Criteria:
    - Domain layer does not talk to the adapter layer
    - App layer handles the logic, domain layer handles transformations and data

### APP-002 Game data handler
Description: Create a handler for game data that fetches game data for tracked games every 5 minutes and determines if the active player has changed or the game has ended.
Acceptance Criteria:
    - Fetch game data on a schedule
    - Add games to tracked list from discord messages
    - Remove games from tracked list that have been completed
    - Check saved game data againt newly scraped data for change

### APP-003 Notification handler
Description: Create an event based notification handler that receives game data change events from the game data handler and handles sending notifications to the correct user in the correct channel and server/guild
Acceptance Criteria:
    - Sends notification on player change to channels if the new current player has signed up for notifications in that channel
    - Sends notification on game end to channels where one or more players were signed up for notifications for that game

### APP-004 User data handler
Description: Create a user data handler that saves and removes mappings between discord users and 18xx account names
Acceptance Criteria:
    - Register new user
    - Link user to 18xx account(s)
    - Remove links to account(s)

### APP-005 Instance handler
Description: Create a handler for new discord servers/guilds and channels that the discord bot may get added to
Acceptance Criteria:
    - Can add a new discord server and channel from messages received by the adapter
    - Can add a new discord channel to existing servers
    - Remove server if bot is removed from a server
    - Remove channel server if bot is removed from a channel
    - Provide data to the notification handler for where to send the notification

## Deployment

### DEP-001 Run dgraph as container
Description: Set up a dgraph container that can be read by the main service
Acceptance Criteria:
    - The 18xxNotifier service can read and write data to the dgraph database

### DEP-002 Initialise the schema on start up via a schema-init container
Description: Set up a initialiser container that waits for dgraph to be ready and sends the schema file at docs/schema.sdl to the dgraph container
Acceptance Criteria:
    - Can initialise the schema in the dgraph container

### DEP-003 Run as a container
Description: Create an image that can run as a container for the 18xxNotifier service
Acceptance Criteria:
    - Docker compose can run the database, database-init and the 18xxNotifier service together
