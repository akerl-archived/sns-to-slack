package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/akerl/go-lambda/cloudwatch"
	"github.com/akerl/go-lambda/s3"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/yaml.v2"
)

var config *configFile

type configFile struct {
	Alarms map[string]alarmConfig `json:"alarms"`
}

func (c *configFile) lookup(name string) (alarmConfig, error) {
	conf, ok := c.Alarms[name]
	if ok {
		return conf, nil
	}
	conf, ok = c.Alarms["default"]
	if ok {
		return conf, nil
	}
	return conf, fmt.Errorf("no config found: %s", name)
}

type alarmConfig struct {
	URL string `json:"url"`
}

func (ac *alarmConfig) notify(m cloudwatch.AlarmMessage) error {
	slack := slackMsg{Text: m.Subject}
	buf, err := json.Marshal(slack)
	if err != nil {
		return err
	}
	_, err = http.Post(ac.URL, "application/json", bytes.NewReader(buf))
	return err
}

type slackMsg struct {
	Text string `json:"text"`
}

func handler(e cloudwatch.SNSEvent) error {
	if config == nil {
		return fmt.Errorf("config failed to load")
	}

	for _, r := range e.Records {
		m, err := r.DecodedMessage()
		if err != nil {
			return err
		}
		ac, err := config.lookup(m.AlarmName)
		if err != nil {
			return err
		}
		err = ac.notify(m)
		if err != nil {
			return err
		}
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

	c := configFile{}
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
