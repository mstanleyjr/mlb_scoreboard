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
Run locally without a Pi by enabling the built-in matrix emulator:

```bash
go run main.go --simulate
```

Terminal-only simulation mode (no desktop window):

```bash
go run main.go --simulate-terminal
```

You can also use env vars instead of flags:

```bash
MLB_SCOREBOARD_SIMULATE=1 go run main.go
MLB_SCOREBOARD_SIMULATE_TERMINAL=1 go run main.go
```

## Raspberry Pi Hardware Run

```bash
sudo -E go run main.go --led-row=64 --led-cols=64 --led-gpio-mapping=adafruit-hat
```

## Custom Terminal Matrix Simulator (No rgbmatrix emulator)
Print a 64x64 matrix directly in terminal using:
- `R`, `G`, `B` for colored pixels
- space (or `-`) for empty pixels

```bash
go run ./cmd/termsim
go run ./cmd/termsim --empty=-
go run ./cmd/termsim --league="AL West"
```

Use this for fast layout tuning on macOS without the Pi hardware path.
