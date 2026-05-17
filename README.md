# mlb_scoreboard
Golang MLB Scoreboard for Raspberry Pi

## Makefile Helpers

```bash
make help
make sim
make sim-term
make pi-run
make fmt
make test
```

## Local 64x64 Simulation
Run the default local terminal simulator:

```bash
go run .
```

Terminal-only simulation mode (no desktop window):

```bash
go run ./cmd/termsim
```

## Raspberry Pi Hardware Run

```bash
sudo -E go run -tags pi .
```

Startup settings are read from `mlb_scoreboard.json` in the repo root by default. You can point at another file with `-config`, and any explicitly passed flags still override the config file.

The config can also control page-specific behavior, including the next matchup V2 renderer:

```json
{
  "division_standings": {
    "monochrome": true,
    "green_background": true
  },
  "next_matchup": {
    "version": "v2",
    "hold_ms": 2500,
    "slide_ms": 500
  }
}
```

Example:

```bash
sudo -E go run -tags pi . -config /path/to/mlb_scoreboard.json
```

## Custom Terminal Matrix Simulator (No rgbmatrix emulator)
Print a 64x64 matrix directly in terminal using:
- `R`, `G`, `B` for colored pixels
- space (or `-`) for empty pixels

```bash
go run ./cmd/termsim
go run ./cmd/termsim --empty=-
go run ./cmd/termsim -config /path/to/mlb_scoreboard.json
```

Use this for fast layout tuning on macOS without the Pi hardware path.
