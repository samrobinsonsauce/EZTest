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

### Releasing

1. Tag a new version:
   ```bash
   git tag v1.0.0
   git push --tags
   ```

2. GoReleaser will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Update the Homebrew formula

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Setting Up Homebrew Tap (For Maintainers)

To enable `brew install samrobinsonsauce/tap/ezt`, you need to set up a Homebrew tap repository:

### 1. Create the Tap Repository

Create a new GitHub repository named `homebrew-tap` at:
`https://github.com/samrobinsonsauce/homebrew-tap`

### 2. Initialize the Repository

```bash
git clone https://github.com/samrobinsonsauce/homebrew-tap.git
cd homebrew-tap
mkdir Formula
touch Formula/.gitkeep
git add .
git commit -m "Initial commit"
git push
```

### 3. Create a Personal Access Token

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Create a new token with:
   - **Repository access**: Select `homebrew-tap` repository
   - **Permissions**: 
     - Contents: Read and write
     - Metadata: Read-only
3. Copy the token

### 4. Add the Token as a Secret

In your `ezt` repository:
1. Go to Settings → Secrets and variables → Actions
2. Add a new repository secret:
   - **Name**: `HOMEBREW_TAP_GITHUB_TOKEN`
   - **Value**: (paste your token)

### 5. Release

When you push a tag (e.g., `git tag v1.0.0 && git push --tags`), the GitHub Action will:
1. Build binaries for macOS (Intel + Apple Silicon) and Linux
2. Create a GitHub release with the binaries
3. Automatically update the Homebrew formula in your tap repository

### 6. Install

Your team can now install with:

```bash
brew tap samrobinsonsauce/tap
brew install ezt
```

Or in one command:

```bash
brew install samrobinsonsauce/tap/ezt
```
