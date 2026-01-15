```
  _____  ___   _   _ _____  _   _ _____ 
/  ___|/ _ \ | \ | /  __ \| | | |  _  |
\ `--./ /_\ \|  \| | /  \/| |_| | | | |
 `--. \  _  || . ` | |    |  _  | | | |
/\__/ / | | || |\  | \__/\| | | \ \_/ /
\____/\_| |_/\_| \_/\____/\_| |_/\___/
```
              

A terminal-first deep work companion for planning, scheduling, and tracking focused sessions.
Built as a fast, local, and distraction-free TUI with a tiny SQLite store.

> "La diligencia es madre de la buena ventura." - Sancho Panza

## Why sancho

- Deep work blocks with clear start/end times
- Fast CLI + TUI workflow for planning and review
- Local-first data with a pure Go SQLite driver (no CGO)
- LMStudio, Ollama and GitHub Copilot integration for LLM-assisted planning

## Quick start

```bash
# build optimized darwin/arm64 binary
make build

# run the app
./bin/sancho
```

## Configuration

Config is layered: defaults -> config file -> env vars.

- Config file: `~/.config/sancho/config.toml`
- Database: `~/.local/share/sancho/sancho.db`

Example snippet:

```toml
[llm]
provider = "copilot"
model = "gpt-4o-mini"
```

## Development

```bash
make build-dev
make test
make lint
```

## Roadmap

- Weekly review dashboards
- Calendar export
- Focus session analytics

---

Stay small, stay deliberate. Focus wins.
