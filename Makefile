.PHONY: build

build:
	go build -o ${HOME}/.local/bin/ipurge ./src/tui/tui.go
