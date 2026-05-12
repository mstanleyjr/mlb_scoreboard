
.PHONY: help sim sim-term pi-run fmt test

help:
	@echo "Targets:"
	@echo "  sim       Run local 64x64 matrix emulator"
	@echo "  sim-term  Run terminal-only matrix emulator"
	@echo "  pi-run    Run on Raspberry Pi hardware"
	@echo "  fmt       Format Go files"
	@echo "  test      Run Go tests/build checks"

sim:
	go run .

sim-term:
	go run ./cmd/termsim

pi-run:
	sudo -E go run -tags pi .

fmt:
	go fmt ./...

test:
	go test ./...
