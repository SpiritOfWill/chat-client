NAME=$(shell basename `pwd`)
VERSION=1.0
REVISION=0
BUILD?=0

build:	$(shell ls *.go)
	go build -o $(NAME) -ldflags "-extldflags '-static' -s -w -X main.appName=$(NAME) -X main.appVers=$(VERSION)-$(BUILD) \
		-X main.appDate=`date -u +%Y-%m-%d.%T`"

run:	
	./$(NAME)
