This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Restreamer Monitor Go is a multi-platform live streaming monitoring and relay tool written in Go. It monitors live streams (currently supporting Bilibili) and can relay streams to multiple destinations using FFmpeg.

## Development Commands

### Building
```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Clean build artifacts
make clean
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running
```bash
# Monitor live streams
./RestreamerMonitor monitor -c config.json -i 30s -v

# Relay live streams
./RestreamerMonitor relay -c config.json -v
```

## Configuration

The application uses JSON configuration (`config.json`) with these main sections:

- **rooms**: List of rooms to monitor with platform, room_id, and enabled status
- **relays**: Stream relay configurations with source, destinations, and FFmpeg options
- **telegram**: Bot configuration with token, chat IDs, admin IDs, and notification settings
- **interval**: Monitoring check interval
- **verbose**: Debug logging flag

## Architecture Notes

### Platform Integration
New platforms can be added by implementing the `StreamSource` interface. The Bilibili implementation (`service/bilibili.go`) serves as a reference.

### Notification System
The notification manager supports multiple channels and types:
- System events (startup, shutdown)
- Monitor events (stream start/stop)
- Relay events (relay status changes)
- Error events (failures and exceptions)

### Telegram Bot Features
Full bot integration with:
- Real-time notifications
- Remote control commands (`/status`, `/rooms`, `/relays`, `/stop`, `/restart`)
- Admin permission system
- Multi-chat support

### Stream Relaying
Uses FFmpeg for high-performance stream processing with:
- Multiple destination support
- Quality settings (720p, 480p, best, worst)
- Custom FFmpeg options per destination
- Process management and restart capabilities

## Tool Usage Guidelines  
This repository recommends using Serena MCP to provide IDE-level code understanding and safe editing capabilities for Claude. The following are conventions for using Serena in this repository.  

- If you don’t understand how to use Serena, first call `initial_instructions` to obtain its official user manual and tool list.  
- When searching for definitions, references, call relationships, type information, or cross-file dependencies, prioritize using Serena’s LSP/symbol-related tools instead of relying solely on full-text search.  
- When refactoring or making bulk modifications across multiple files, first use Serena to retrieve a list of affected symbols and references, then create a modification plan and apply editing actions in batches.  
- Avoid reading particularly large files all at once. Instead, let Serena precisely jump to the location of relevant functions, classes, interfaces, or type definitions to reduce interference from irrelevant context.  
- For tasks like "explaining architecture and organizing module relationships," first use Serena to obtain the project structure, dependencies, and call graphs, then output structured analysis based on this information.  
- When issues involve API, configuration, best practices, or version differences of third-party libraries, prioritize using documentation query tools like `get-library-docs` to answer based on the latest official documentation rather than relying on subjective memory.  

### Serena Memory Tools and Collaboration Guidelines  
To ensure the sustainable transfer of project context, collaboration knowledge, and historical decisions, the following memory management conventions must be followed:  

- All long-term collaboration knowledge, architecture specifications, and phase progress must be synchronized to separate `write_memory` files (e.g., `project_guidelines.md`, `feature_progress.md`).  
- Before entering a new session or switching tasks, always check available historical memories with `list_memories` and invoke `read_memory` to reuse knowledge.  
- Important decisions/changes must be archived with `write_memory` immediately after coding. Synchronize memory progress automatically when a phase-specific task ends or the development pipeline is interrupted, to enable subsequent tracking and redevelopment.  
- For erroneous or outdated memory content, prioritize correcting it with `edit_memory` rather than adding duplicate memories. 
- For outdated or useless memory content, invoke `delete_memory` to clean it up to avoid misleading future collaboration.  

## Task Workflow Recommendations  

- Before starting a new task, first clarify the task objectives (bug fix, feature addition, refactoring, documentation) and summarize the expected results in 1–3 sentences.  
- Use Serena to retrieve a list of files, key symbols, and dependencies related to the task to avoid editing without understanding the context.  
- For medium- to large-scale changes, first provide a "list of planned steps" (analysis, localization, modification, testing, documentation update, etc.) and then execute them one by one, providing a brief summary at the end of each step.  
- After completing important changes, it is recommended to run build or test scripts and check for type errors, runtime errors, or regression risks in the output.  

## Safety and Restrictions  

- Prohibit performing irreversible operations (deleting directories, resetting Git repositories, overwriting configuration files, etc.) without explicit authorization.  
- Only make modifications to files, install new dependencies, or edit configuration files when necessary and recoverable, and whenever possible, isolate changes using Git branches.  
- If uncertain about the safety or impact of an action, first provide a solution in the form of "analysis and recommendations" rather than executing it directly.