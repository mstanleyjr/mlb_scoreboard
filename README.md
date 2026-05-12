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
