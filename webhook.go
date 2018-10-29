package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)


const constUrl = " "


type UrlHook struct {
	Type 	string		`json:"type"`
}


type TriggerValue struct {
	Type 	int				`json:"value"`
}


type WebhookInfo struct {
	WebhookURL 				string 	`json:"webhookURL"`
	MinTriggerValue 	int			`json:"minTriggerValue"`
}


func getWebhook(what string) {
	temp := WebhookInfo{}

	temp.WebhookURL = " "

	raw, _ := json.Marshal(info)

	resp, err := http.Post(constUrl, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println(err)
		fmt.Println(ioutil.ReadAll(resp.Body))
	}
}

func main() {
	for {
		text := "Heroku timer test at: " + time.Now().String()
		delay := time.Minute * 15

		getWebhook(text)
		time.Sleep(delay)
	}
}
