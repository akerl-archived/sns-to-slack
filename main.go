package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akerl/go-lambda/s3"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/yaml.v2"
)

// TODO: Add config file structure
type configFile struct{}

var config *configFile

func handler(e events.SNSEvent) error {
	log.Printf("%+v", e)
	if config == nil {
		return fmt.Errorf("Config failed to load")
	}
	return nil
}

func loadConfig() {
	bucket := os.Getenv("S3_BUCKET")
	path := os.Getenv("S3_KEY")
	if bucket == "" || path == "" {
		log.Print("variables not provided")
		return
	}

	obj, err := s3.GetObject(bucket, path)
	if err != nil {
		log.Print(err)
		return
	}

	c := config{}
	err = yaml.Unmarshal(obj, &c)
	if err != nil {
		log.Print(err)
		return
	}
	config = &c
}

func main() {
	loadConfig()
	lambda.Start(handler)
}
