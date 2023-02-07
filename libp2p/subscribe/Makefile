.PHONY: all build
FILE=libp2p-subscribe

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
