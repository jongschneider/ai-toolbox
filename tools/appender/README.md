# Appender

Appender is a terminal-based file explorer and content collector tool that allows you to navigate your filesystem, select files, and export their content into a single file or to the clipboard.

## Features

- Interactive file explorer with tree view
- File selection and preview
- Hidden files toggling
- Content export to file or clipboard
- File search with glob pattern support
- Keyboard-driven navigation

## Installation

```bash
go install github.com/jongschneider/ai-toolbox/tools/appender@latest
```

Or build from source:

```bash
git clone https://github.com/jongschneider/ai-toolbox.git
cd ai-toolbox/tools/appender
go build
```

## Usage

```bash
appender [directory]
```

If no directory is specified, the current directory will be used.

### Configuration

Configuration is handled through environment variables and command-line flags:

- `APPENDER_LOGGING`: Set logging level (1=DEBUG, 2=INFO, 3=WARN, 4=ERROR)
- `-l, --logging`: Alternative way to set logging level via command line

Example:
```bash
APPENDER_LOGGING=2 appender ./my-project
# Or using flags
appender -l 2 ./my-project
```

## Controls

### Basic Navigation
- `↑/k`: Move cursor up
- `↓/j`: Move cursor down
- `l`: Expand directory
- `h`: Collapse directory
- `Home`: Jump to top
- `End`: Jump to bottom
- `PgUp`: Move cursor up one page
- `PgDown`: Move cursor down one page

### File Operations
- `Space`: Select/deselect file or directory
- `Enter`: Save selected files to output file
- `c`: Copy selected files to clipboard
- `.`: Toggle hidden files
- `q` or `Ctrl+C`: Quit application

### Search Operations
- `/`: Enter search mode
- `Enter` (in search mode): Execute search and exit search mode
- `Esc` (in search mode): Exit search mode
- `n`: Navigate to next match
- `N`: Navigate to previous match

### Content Viewing
- `K`: Scroll preview pane up 
- `J`: Scroll preview pane down
- `g`: Scroll preview pane to top
- `G`: Scroll preview pane to bottom

### Help
- `?`: Toggle help view

## Search Functionality

The search feature uses glob patterns to find files and directories in your workspace:

- Type `/` to enter search mode
- Enter a glob pattern (e.g., `*.go`, `**/*.md`)
- Press `Enter` to execute the search
- Use `n` and `N` to navigate between matches
- Press `Esc` to clear search results

Glob patterns support:
- `*`: Matches any sequence of characters in a filename
- `**`: Matches any sequence of directories 
- `?`: Matches any single character
- `[abc]`: Matches any of the characters in brackets
- `[!abc]`: Matches any character not in brackets

Examples:
- `*.go`: All Go files in the current directory
- `**/*.go`: All Go files in any subdirectory
- `cmd/*/main.go`: All main.go files one level under the cmd directory
- `[a-c]*/`: All directories starting with a, b, or c

The search respects your hidden files setting, so `.git` directories will be excluded when hidden files are toggled off.

## Output Format

The output file will contain the content of all selected files, with each file's content preceded by a comment line containing the file path.

Example:
```
# path/to/file1.go
package main

func main() {
    // ...
}

# path/to/file2.go
package utils

func Utility() {
    // ...
}
```

## Filtering

Appender automatically filters binary files and can toggle the visibility of hidden files (files and directories starting with `.`).

## Troubleshooting

Logs are written to `./logs/debug.log` when logging is enabled. Increase the logging level for more detailed information.
