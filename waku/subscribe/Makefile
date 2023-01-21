.PHONY: all build
FILE=waku-subscribe

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
