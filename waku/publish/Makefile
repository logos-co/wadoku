.PHONY: all build
FILE=waku-publish

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
