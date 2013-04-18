package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Status struct {
	Events []Event `json:"events"`
}

type Event struct {
	Id      int      `json:"event_id"`
	Message *Message `json:"message"`
}

type Message struct {
	Id              string `json:"id"`
	Room            string `json:"room"`
	PublicSessionId string `json:"public_session_id"`
	IconUrl         string `json:"icon_url"`
	Type            string `json:"type"`
	SpeakerId       string `json:"speaker_id"`
	Nickname        string `json:"nickname"`
	Text            string `json:"text"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var status Status
			err := json.NewDecoder(r.Body).Decode(&status)
			if err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			results := ""
			for _, event := range status.Events {
				tokens := strings.SplitN(event.Message.Text, " ", 2)
				if len(tokens) == 1 && tokens[0] == "!heroku" {
					results += "へろくー"
				} else if len(tokens) == 2 && tokens[0] == "!weather" {
					url := fmt.Sprintf("http://openweathermap.org/data/2.1/find/name?q=%s", tokens[1])
					if res, err := http.Get(url); err == nil {
						defer res.Body.Close()
						if res.StatusCode == 200 {
							var weather map[string]interface{}
							if json.NewDecoder(res.Body).Decode(&weather) == nil {
								icon := weather["list"].([]interface{})[0].(map[string]interface{})["weather"].([]interface{})[0].(map[string]interface{})["icon"].(string)
								results += fmt.Sprintf("http://openweathermap.org/img/w/%s.png", icon) + "\n"
							}
						}
					}
				}
			}
			if len(results) > 0 {
				results = strings.TrimRight(results, "\n")
				if runes := []rune(results); len(runes) > 1000 {
					results = string(runes[0:999])
				}
				fmt.Fprintln(w, results)
			}
		} else {
			fmt.Fprintln(w, "こんにちわ世界 "+time.Now().String())
		}
	})
	fmt.Println("listening...")
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
