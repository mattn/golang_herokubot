package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type Weather struct {
	List []struct {
		Weather []struct {
			Icon string `json:"icon"`
		} `json:"weather"`
	} `json:"list"`
}

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

// TODO(ymotonpoo): Replace this by actual struct.
type Fortune map[string]interface{}

func (f Fortune) star(key string) string {
	n := int(f[key].(float64)) // TODO(ymotongpoo): Add type validation.
	star := ""
	for i := 0; i < n; i++ {
		star += "★"
	}
	for i := 0; i < 5-n; i++ {
		star += "☆"
	}
	return star
}

func (f Fortune) IsSign(s string) bool {
	sign := f["sign"].(string)
	if sign == s {
		return true
	}
	return false
}

func (f Fortune) Write(w io.Writer) error {
	data := struct {
		Rank    int
		Total   string
		Love    string
		Money   string
		Job     string
		Color   string
		Item    string
		Sign    string
		Content string
	}{
		Rank:    int(f["rank"].(float64)),
		Total:   f.star("total"),
		Love:    f.star("love"),
		Money:   f.star("money"),
		Job:     f.star("job"),
		Color:   f["color"].(string),
		Item:    f["item"].(string),
		Sign:    f["sign"].(string),
		Content: f["content"].(string),
	}
	tmplText := `{{ .Rank }}位 {{ .Sign }}
総合: {{ .Total }}
恋愛運： {{ .Love }}
金運: {{ .Money }}
仕事運: {{ .Job }}
ラッキーカラー: {{ .Color }}
ラッキーアイテム: {{ .Item }}
{{ .Content }}
`
	tmpl := template.Must(template.New("fortune").Parse(tmplText))
	err := tmpl.Execute(w, data)
	return err
}

func handleEvents(events []Event) string {
	results := ""
	for _, event := range events {
		tokens := strings.SplitN(event.Message.Text, " ", 2)
		if len(tokens) == 1 && tokens[0] == "!heroku" {
			results += "へろくー"
		} else if len(tokens) == 2 && tokens[0] == "!weather" {
			url := fmt.Sprintf("http://openweathermap.org/data/2.1/find/name?q=%s", tokens[1])
			if res, err := http.Get(url); err == nil {
				defer res.Body.Close()
				if res.StatusCode == 200 {
					var weather Weather
					if json.NewDecoder(res.Body).Decode(&weather) == nil {
						icon := weather.List[0].Weather[0].Icon
						results += fmt.Sprintf("http://openweathermap.org/img/w/%s.png", icon) + "\n"
					}
				}
			}
		} else if len(tokens) == 2 && tokens[0] == "!uranai" {
			key := time.Now().Format("2006/01/02")
			url := fmt.Sprintf("http://api.jugemkey.jp/api/horoscope/free/%s", key)
			if res, err := http.Get(url); err == nil {
				defer res.Body.Close()
				if res.StatusCode == 200 {
					horoscope := make(map[string]interface{})
					decoder := json.NewDecoder(res.Body)
					err := decoder.Decode(&horoscope)
					if err != nil {
						results += "ぬっこわれたー"
					} else {
						data := horoscope["horoscope"].(map[string]interface{})
						for _, d := range data[key].([]interface{}) {
							f := Fortune(d.(map[string]interface{}))
							if f.IsSign(tokens[1]) {
								var text bytes.Buffer
								err = f.Write(&text)
								if err != nil {
									results += "ぬっこわれたー"
								}
								results += text.String()
								break
							}
						}
					}
				}
			}

		}
	}
	return results
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
			results := handleEvents(status.Events)
			if len(results) > 0 {
				results = strings.TrimRight(results, "\n")
				if runes := []rune(results); len(runes) > 1000 {
					results = string(runes[0:999])
				}
				fmt.Fprintln(w, results)
			}
		} else {
			fmt.Fprintln(w, "こんにちわ世界ナリー "+time.Now().String())
		}
	})
	fmt.Println("listening...")
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
