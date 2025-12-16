modules:
	go mod tidy

build:
	go build -o plr ./...

run-example: build
	cat ./playground/log.json | PRETTY_LOGRUS_HOME=~/.config/eaardal/plr-debug ./plr
