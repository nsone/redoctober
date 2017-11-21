package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type SlackClient struct {
	Url string
}

func NewSlackClient(url string) *SlackClient {
	return &SlackClient{url}
}

func (slack *SlackClient) Notify(msg, color string) error {
	if slack.Url == "" {
		return errors.New("URL unset")
	}

	msgBody := map[string]interface{}{
		"text": msg,
	}
	body, err := json.Marshal(msgBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", slack.Url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	Client := http.Client{}
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("Could not post to Slack for the reason %s", err.Error())
		return err
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not post to Slack for the reason %s", err.Error())
		return err
	}
	return nil
}
