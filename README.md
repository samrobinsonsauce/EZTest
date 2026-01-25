# ezt - Elixir Test Selector

ezt is a fast, minimal TUI for discovering, selecting, and running Elixir test files from inside a Phoenix or Elixir project. It scans your project for `*_test.exs` files, lets you filter and select any number of them, then runs `mix test` only for the tests you chose. Your selections are saved per project so you can pick up where you left off.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## What it does

- Finds the Elixir project root by walking upward until it detects `mix.exs`
- Lists all `*_test.exs` files it can find (including common umbrella layouts)
- Lets you filter the list with fuzzy search and select multiple files
- Runs `MIX_ENV=test mix test` with only the selected test file paths
- Persists selections per project so the next run starts pre-selected

## Features

- Fuzzy search: quickly filter test files as you type
- Multi-select: select any number of tests to run
- Persistent selections: remembers selections per project
- Vim-style navigation: use home row to navigate
- Clean UI: simple, fzf-inspired interface

## Installation

### Homebrew (macOS/Linux)

```bash
brew install samrobinsonsauce/tap/ezt
```

### From source

```bash
go install github.com/samrobinsonsauce/eztest@latest
```

### Manual download

Download the latest release from the releases page:  
https://github.com/samrobinsonsauce/eztest/releases

## Usage

From inside an Elixir/Phoenix project:

```bash
ezt        # Open the TUI to select and run tests
ezt -r     # Run previously saved tests directly (skip the TUI)
```

If you run `ezt` outside an Elixir project, it will fail with an error because it cannot locate `mix.exs`.

## Key bindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor up/down |
| `↑` / `↓` | Move cursor up/down |
| `Tab` | Toggle selection on the current item |
| `Ctrl+a` | Select all visible (filtered) items |
| `Ctrl+d` | Deselect all items |
| `Enter` | Save selections and run `mix test` for selected files |
| `Ctrl+s` | Save selections and quit (without running) |
| `Esc` | Quit without saving |

## Search

Start typing to filter the test files. Search supports multi-word fuzzy matching, so typing:

```
user controller
```

will match paths that contain both terms in any order.

## Persistent selections

Selections are stored per project under your user config directory. On macOS and Linux this is typically:

```
~/.config/ezt/
```

When you run `ezt` again in the same project, previously selected tests are pre-selected.

## Requirements

- An Elixir project containing `mix.exs`
- `mix` available on your PATH
- Test files under `test/` (and for umbrella projects, under `apps/*/test/`)

## Contributing

Contributions are welcome.

### Development setup

1. Fork the repository and clone your fork.
2. Ensure you have a recent Go toolchain installed.
3. Install dependencies, run tests, and start the app:

```bash
go mod tidy
go test ./...
go run .
```

### Making changes

- Keep changes focused and small where possible.
- Prefer clear names and straightforward control flow over clever abstractions.
- If you add a new feature or fix a bug, include tests where it makes sense.

### Commit and PR guidelines

- Create a feature branch from `main`.
- Use descriptive commit messages.
- Open a pull request with:
  - What changed and why
  - How to test the change
  - Screenshots or a short recording for UI changes (optional but helpful)

## License

MIT License - see [LICENSE](LICENSE) for details.
