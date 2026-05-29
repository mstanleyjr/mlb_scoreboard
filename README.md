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

The config can also control hardware and page-specific behavior, including matrix brightness plus live game, next matchup, and last matchup timing:

```json
{
  "division_standings": {
    "monochrome": true,
    "green_background": true
  },
  "hardware": {
    "brightness": 100
  },
  "live_game": {
    "hold_ms": 2200,
    "slide_ms": 400
  },
  "live_look_in": {
    "repeats": 1
  },
  "next_matchup": {
    "hold_ms": 2500,
    "slide_ms": 500
  },
  "last_matchup": {
    "hold_ms": 2500,
    "slide_ms": 500
  }
}
```

Example:

```bash
sudo -E go run -tags pi . -config /path/to/mlb_scoreboard.json
```

## Scheduled Raspberry Pi Run

For a board that should turn on and off at specific times, use `systemd` rather than cron:

- run the scoreboard as a `systemd` service
- use one timer to start it
- use one timer to stop it

This is more reliable than cron, survives reboots cleanly, and gives you logs through `journalctl`.

### 1. Build a binary

```bash
go build -tags pi -o mlb_scoreboard .
```

### 2. Create the service

Example `/etc/systemd/system/mlb-scoreboard.service`:

```ini
[Unit]
Description=MLB Scoreboard
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/home/stanley/Projects/mlb_scoreboard
ExecStart=/home/stanley/Projects/mlb_scoreboard/mlb_scoreboard -config /home/stanley/Projects/mlb_scoreboard/mlb_scoreboard.json
Restart=on-failure
User=stanley

[Install]
WantedBy=multi-user.target
```

### 3. Create a start timer

Use a oneshot helper service so the timer can start the main service:

`/etc/systemd/system/mlb-scoreboard-start.service`

```ini
[Unit]
Description=Start MLB Scoreboard

[Service]
Type=oneshot
ExecStart=/bin/systemctl start mlb-scoreboard.service
```

`/etc/systemd/system/mlb-scoreboard-start.timer`

```ini
[Unit]
Description=Start MLB Scoreboard on schedule

[Timer]
OnCalendar=*-*-* 15:30:00
Persistent=true

[Install]
WantedBy=timers.target
```

### 4. Create a stop timer

`/etc/systemd/system/mlb-scoreboard-stop.service`

```ini
[Unit]
Description=Stop MLB Scoreboard

[Service]
Type=oneshot
ExecStart=/bin/systemctl stop mlb-scoreboard.service
```

`/etc/systemd/system/mlb-scoreboard-stop.timer`

```ini
[Unit]
Description=Stop MLB Scoreboard on schedule

[Timer]
OnCalendar=*-*-* 23:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

### 5. Enable everything

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now mlb-scoreboard-start.timer
sudo systemctl enable --now mlb-scoreboard-stop.timer
```

### Controls

Use these commands on the Pi to manage the service and timers:

```bash
sudo systemctl start mlb-scoreboard.service
sudo systemctl stop mlb-scoreboard.service
sudo systemctl restart mlb-scoreboard.service
sudo systemctl status mlb-scoreboard.service

sudo systemctl start mlb-scoreboard-start.timer
sudo systemctl stop mlb-scoreboard-start.timer
sudo systemctl status mlb-scoreboard-start.timer

sudo systemctl start mlb-scoreboard-stop.timer
sudo systemctl stop mlb-scoreboard-stop.timer
sudo systemctl status mlb-scoreboard-stop.timer

sudo systemctl list-timers --all
```

Adjust the `OnCalendar` values to match the times you want the board on and off.

### Logs

The scoreboard logs go to the `systemd` journal by default.

```bash
sudo journalctl -u mlb-scoreboard.service
sudo journalctl -u mlb-scoreboard.service -f
```

`-f` means **follow**, so it stays attached and streams new log lines live.

If you later add more `systemd` units, you can inspect them the same way:

```bash
sudo journalctl -u mlb-scoreboard-start.service
sudo journalctl -u mlb-scoreboard-stop.service
```

By default, journal logs are stored under one of these locations on the Pi:

- `/var/log/journal` when persistent journaling is enabled
- `/run/log/journal` when logs are only kept until reboot

### Updating the app manually

```bash
cd /home/stanley/Projects/mlb_scoreboard
git pull --ff-only
go build -tags pi -o mlb_scoreboard .
sudo systemctl restart mlb-scoreboard.service
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
