build:
	GOOS=linux go build -o bin/main

deploy:
	sls deploy

run:
	sls invoke -f hatenaScraping