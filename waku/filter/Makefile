.PHONY: all build
FILE=waku-filter

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
