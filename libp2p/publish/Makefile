.PHONY: all build
FILE=libp2p-publish

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
