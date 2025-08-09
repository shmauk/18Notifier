## Overview

This document outlines the technical architecture for a go microservice. The system follows a hexagonal architecture with event-driven and routine based communication patterns.

## Technology Stack

- **Database**: Dgraph
- **Language**: Go
- **Deployment**: Docker

## Communication

- **18xx API**: Pull from the api for tracked games every 5 minutes
- **Discord Commands**: Commands received from discord are events handled by the microservice
- **Discord Notifications**: Notifications are saved to the database and the unsent notifications are sent every minute