# ezt - Elixir Test Selector

A fast, and minimal TUI for selecting and running Elixir tests. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![ezt demo](https://via.placeholder.com/800x400?text=ezt+demo)

## Features

-  **Fuzzy search** - quickly filter test files as you type
-  **Multi-select** - select any number of tests to run
-  **Persistent selections** - remembers your selections per project
- ️ **Vim-style navigation** - use `j`/`k` to move up and down
-  **Clean UI** - clean, fzf-inspired interface

## Installation

### Homebrew (macOS/Linux)

```bash
brew install samrobinsonsauce/tap/ezt
```

### From Source

```bash
go install github.com/samrobinsonsauce/ezt@latest
```

### Manual Download

Download the latest release from the [releases page](https://github.com/samrobinsonsauce/ezt/releases).

## Usage

Navigate to your Elixir/Phoenix project and run:

```bash
ezt        # Open TUI to select and run tests
ezt -r     # Run previously saved tests directly (skip TUI)
```

### Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `Ctrl+k` | Move cursor up |
| `↓` / `Ctrl+j` | Move cursor down |
| `Tab` | Toggle selection on current item |
| `Ctrl+a` | Select all visible (filtered) items |
| `Ctrl+d` | Deselect all items |
| `Enter` | Run selected tests with `mix test` |
| `Ctrl+s` | Save selections and quit (without running) |
| `Esc` | Quit without saving |

### Search

Just start typing to filter the test files. The search supports **multi-word fuzzy matching** — type `user controller test` to find files containing all those words in any order.

### Persistent Selections

When you press `Enter`, your selections are saved to `~/.config/ezt/state.json`. The next time you run `ezt` in the same project, your previous selections will be restored.

## Requirements

- An Elixir project with a `test/` directory
- `mix` available in your PATH

## Configuration

Selections are stored per-project in `~/.config/ezt/state.json`.

## Development

### Building

```bash
go build -o ezt .
```

### Running locally

```bash
go run .
```

## License

MIT License - see [LICENSE](LICENSE) for details.

---

