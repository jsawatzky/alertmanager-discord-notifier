package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type AlertNotification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type DiscordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Url         string              `json:"url,omitempty"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Discord color values
const (
	ColorRed   = 0x992D22
	ColorGreen = 0x2ECC71
	ColorGrey  = 0x95A5A6
)

// Discord emojis
const (
	EmojiFiring   = ":bangbang:"
	EmojiResolved = ":white_check_mark:"
	EmojiUnknown  = ":grey_question:"
)

// Discord empty value
var (
	Empty           = "\u200b"
	EmptyEmbedField = DiscordEmbedField{
		Name:  Empty,
		Value: Empty,
	}
)

var debug bool

func processAlert(a AlertNotification) error {
	alertColor := ColorGrey
	alertEmoji := EmojiUnknown

	switch a.Status {
	case "firing":
		alertColor = ColorRed
		alertEmoji = EmojiFiring
	case "resolved":
		alertColor = ColorGreen
		alertEmoji = EmojiResolved
	}

	labelField := DiscordEmbedField{
		Name:   "Labels",
		Value:  "",
		Inline: true,
	}
	for l, v := range a.CommonLabels {
		labelField.Value += fmt.Sprintf(" - %s = %s\n", l, v)
	}

	annotationField := DiscordEmbedField{
		Name:   "Annotations",
		Value:  "",
		Inline: true,
	}
	for l, v := range a.CommonAnnotations {
		annotationField.Value += fmt.Sprintf(" - %s = %s\n", l, v)
	}

	dm := DiscordMessage{
		Embeds: []DiscordEmbed{
			{
				Title:  fmt.Sprintf("%s [%s:%d] %s", alertEmoji, strings.ToUpper(a.Status), len(a.Alerts), a.CommonLabels["alertname"]),
				Url:    a.ExternalURL,
				Color:  alertColor,
				Fields: []DiscordEmbedField{labelField, annotationField},
			},
		},
	}

	dmBody, err := json.Marshal(dm)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	if debug {
		log.Printf("[DEBUG] sending discord webhook request with body: %s", dmBody)
	}

	resp, err := http.Post(viper.GetString("webhook"), "application/json", bytes.NewReader(dmBody))
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	} else if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send webhook request: %s", resp.Status)
	}

	return nil
}

func listenForAlerts() error {
	listenAddr := viper.GetString("listen")
	log.Printf("Listening for alerts on %s", listenAddr)
	return http.ListenAndServe(listenAddr, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("received %s on %s/%s", r.Method, r.Host, r.URL.RawPath)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ERROR: failed to read request body: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if debug {
			fmt.Printf("[DEBUG] received: %s", body)
		}

		a := AlertNotification{}
		err = json.Unmarshal(body, &a)
		if err != nil {
			fmt.Printf("ERROR: failed to unmarshal request body: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		err = processAlert(a)
		if err != nil {
			log.Printf("ERROR: failed to process alert: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println("successfully processed alert")
		rw.WriteHeader(http.StatusNoContent)
	}))
}

func main() {
	viper.SetEnvPrefix("adn")
	viper.AutomaticEnv()

	pflag.StringP("webhook", "w", "", "Discord webhook URL")
	pflag.StringP("listen", "l", "0.0.0.0:9094", "<address>:<port> to listen on")
	pflag.BoolP("debug", "d", false, "enable debug logging")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if !viper.IsSet("webhook") {
		log.Fatalln("ERROR: webhook url is not defined")
	}

	debug = viper.GetBool("debug")

	err := listenForAlerts()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Shutting down")
	} else {
		log.Fatalf("ERROR: failed to listen on HTTP: %v", err)
	}
}
