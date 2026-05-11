SHELL := /bin/zsh

.PHONY: help sim sim-term pi-run fmt test

help:
	@echo "Targets:"
	@echo "  sim       Run local 64x64 matrix emulator"
	@echo "  sim-term  Run terminal-only matrix emulator"
	@echo "  pi-run    Run on Raspberry Pi hardware"
	@echo "  fmt       Format Go files"
	@echo "  test      Run Go tests/build checks"

sim:
	go run main.go --simulate

sim-term:
	go run main.go --simulate-terminal

pi-run:
	sudo -E go run main.go --led-row=64 --led-cols=64 --led-gpio-mapping=adafruit-hat

fmt:
	go fmt ./...

test:
	go test ./...

