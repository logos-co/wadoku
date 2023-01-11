.PHONY: all build
FILE=waku-lightpush

build:
	go build -o $(FILE)

docker: build
	docker build --rm -t $(FILE):alpha .

all: build
