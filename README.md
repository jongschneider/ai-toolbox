# AI Toolbox

AI Toolbox is a collection of command-line tools designed to enhance developer productivity when working with Large Language Models (LLMs) and AI systems. These tools streamline common workflows and make it easier to interact with AI services in your development environment.

## Tools

### Appender

Appender is a CLI tool that helps you interact with Chat LLMs like Claude.ai more effectively by making it easy to capture and format code from your local filesystem. It provides an interactive tree-view interface that allows you to select files and directories, then generates a formatted output suitable for pasting into AI chat interfaces.

#### Features

- Interactive file tree navigation with vim-like keybindings
- Visual selection of files and directories
- Automatic file content capture and formatting
- Support for hidden file filtering
- Tree-view collapsing/expanding
- Terminal-based UI with scroll support

#### Installation

Using Nix:

```shell
nix build github:jongschneider/ai-toolbox#appender
```

#### Run in shell

- Run `appender` directly with:

```shell
$ nix run github:jongschneider/ai-toolbox#appender
```

- Run `appender` in new shell with:

```shell
$ nix shell github:jongschneider/ai-toolbox#appender
$ appender
```

Or install from source:

```shell
git clone https://github.com/jongschneider/ai-toolbox.git
cd ai-toolbox
just build appender
```

#### Usage

1. Run the tool in your project directory:

   ```shell
   $ appender
   ```

   Or specify a different directory:

   ```shell
   $ appender /path/to/directory
   ```

2. Navigate the file tree:

   - `j` or `↓`: Move cursor down
   - `k` or `↑`: Move cursor up
   - `space`: Select/deselect file or directory
   - `l` or `h`: Expand/collapse directory
   - `.`: Toggle hidden files
   - `enter`: Generate output
   - `q` or `ctrl+c`: Quit

3. After selecting files and pressing enter, Appender will create an `output.txt` file containing the formatted content of all selected files, ready to be shared with an LLM.

## Development

This project uses Nix flakes for development and building. Make sure you have Nix installed with flakes enabled.

### Prerequisites

- Nix package manager with flakes enabled
- Go 1.23.4 or later

### Development Commands

The repository includes a `justfile` with common development commands:

```shell
# List available commands
just

# Build a specific tool
just build appender

# Run a specific tool
just run appender

# Run tests for a specific tool
just test appender

# Run linter for a specific tool
just lint appender

# Create a new tool
just new-tool toolname
```

### Project Structure

```
.
├── flake.nix           # Nix flake configuration
├── go.work            # Go workspace configuration
├── justfile           # Development command definitions
└── tools/            # Directory containing all tools
    └── appender/     # Appender tool source code
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. Make sure to:

1. Follow the existing code style
2. Add tests for new functionality
3. Update documentation as needed
4. Run tests and linting before submitting (`just test-all && just lint-all`)

## License

[Add your chosen license here]
