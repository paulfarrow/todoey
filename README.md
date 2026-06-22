# Todoey

A terminal-based Todoist client built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). Todoey provides a fast, keyboard-driven interface for managing your Todoist tasks without leaving the terminal.

## Features

- View today's tasks and overdue items, grouped by project
- Browse tasks by project
- Complete, delete, reschedule, and move tasks between projects
- Create tasks using Todoist's quick-add syntax (natural language dates, `#project` tags)
- Search across all tasks
- Visual selection mode for bulk operations
- Task detail view with inline editing
- Auto-refresh on a configurable interval
- Add tasks from the command line without opening the TUI

## Usage

```
todoey                        # Launch the TUI
todoey -a "Buy milk tomorrow" # Add a task and exit
todoey --add-task "Meeting #Work Friday 3pm"
```

## Configuration

Config file location: `$XDG_CONFIG_HOME/todoey/config.json` (defaults to `~/.config/todoey/config.json`)

A default config is created on first run.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `api_token` | string | `""` | Your Todoist API token |
| `auto_refresh` | bool | `true` | Automatically refresh task list |
| `refresh_interval_seconds` | int | `60` | Seconds between auto-refreshes |

The `TODOIST_API_TOKEN` environment variable takes precedence over the config file's `api_token` value.

## Hotkeys

### Main View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `{` | Jump to previous project group (Today view) |
| `}` | Jump to next project group (Today view) |
| `J` | Next project |
| `K` | Previous project |
| `T` | Go to Today view |
| `x` | Complete task(s) |
| `d` | Delete task(s) (with confirmation) |
| `a` | Add new task |
| `r` | Reschedule task(s) |
| `/` | Search tasks |
| `c` | Go to project by name |
| `W` | Open task in web browser |
| `Alt+m` | Move task(s) to another project |
| `Alt+r` | Manual refresh |
| `O` | Toggle overdue-only filter |
| `v` / `Space` | Toggle selection on current task |
| `V` | Enter visual selection mode |
| `Enter` | Open task detail view |
| `Esc` | Clear search/filter/selection |
| `q` | Quit (with confirmation) |

Movement keys (`j`, `k`, `J`, `K`) accept a numeric prefix for repeated movement, e.g. `5j` moves down 5 tasks, `3K` moves up 3 projects.

### Visual Mode

| Key | Action |
|-----|--------|
| `j` / `k` | Extend selection up/down |
| `x` | Complete selected tasks |
| `d` | Delete selected tasks |
| `Alt+m` | Move selected tasks |
| `r` | Reschedule selected tasks |
| `V` / `Esc` | Exit visual mode |

### Detail View

| Key | Action |
|-----|--------|
| `e` | Edit task content |
| `E` | Edit task description |
| `r` | Reschedule task |
| `x` | Complete task |
| `d` | Delete task (with confirmation) |
| `Alt+m` | Move task to another project |
| `q` / `Esc` | Back to main view |

### Text Input

| Key | Action |
|-----|--------|
| `Enter` | Confirm input |
| `Esc` | Cancel input |
| `Tab` | Autocomplete project name |
| `←` / `→` | Move cursor |
| `Ctrl+←` / `Ctrl+→` | Move cursor by word |
| `Home` / `Ctrl+a` | Move to start |
| `End` / `Ctrl+e` | Move to end |
| `Backspace` | Delete character before cursor |
| `Ctrl+Backspace` | Delete word before cursor |
| `Ctrl+d` / `Delete` | Delete character after cursor |
