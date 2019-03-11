package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	scraping "go-hatena/src"
)

type Response struct {
	Message string `json:"message"`
}

func Handler() (Response, error) {
	scraping.Run()

	return Response{
		Message: "本日のスクレイピング終了",
	}, nil
}

func main() {
	lambda.Start(Handler)
}
