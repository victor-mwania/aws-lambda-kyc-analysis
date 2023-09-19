package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"kyc-documents-analysis-lambda/handler"
)

func main() {
	lambda.Start(handler.Handler)
}
