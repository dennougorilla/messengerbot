package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"strconv"
	"time"

	. "github.com/uzimaru0000/messengerbot/models"
	"github.com/uzimaru0000/messengerbot/template"
)

var accessToken = os.Getenv("MESSENGERBOT_TOKEN")
var verifyToken = "3460"

const (
	EndPoint = "https://graph.facebook.com/v2.6/me/messages"
)

type TalkJson struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Results []talkApiResult `json:"results"`
}

type talkApiResult struct {
	Perplexity float64 `json:"perplexity"`
	Reply      string  `json:"reply"`
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:9000")
		if err != nil {
			return
		}
	http.HandleFunc("/", TopPageHandler)
	http.HandleFunc("/webhook", webhookHandler)
		fcgi.Serve(l, nil)
}


func TopPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is go-bot application's top page.")
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		verifyTokenAction(w, r)
	}
	if r.Method == "POST" {
		webhookPostAction(w, r)
	}
}

func verifyTokenAction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("hub.verify_token") == verifyToken {
		log.Print("verify token success.")
		fmt.Fprintf(w, r.URL.Query().Get("hub.challenge"))
	} else {
		log.Print("Error: verify token failed.")
		fmt.Fprintf(w, "Error, wrong validation token")
	}
}

func webhookPostAction(w http.ResponseWriter, r *http.Request) {
	var receivedMessage ReceivedMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
	}

	if err = json.Unmarshal(body, &receivedMessage); err != nil {
		log.Print(err)
	}
	messagingEvents := receivedMessage.Entry[0].Messaging
	for i, event := range messagingEvents {
		senderID := event.Sender.ID
		log.Print(i)
		log.Print(event)
		if event.Message != nil {
			if event.Message.Text == "QR" {
				q := []QuickReplies{
					{ContentType: "text", Title: "a", Payload: "a", ImageURL: "https://user-images.githubusercontent.com/28649418/45468742-385b0500-b761-11e8-879e-2a5cef3b8ddc.png"},
					{ContentType: "location"},
					{ContentType: "user_phone_number"},
					{ContentType: "user_email"},
				}
				sendQuickReplies(senderID, "QuickReplies", q)
			} else if event.Message.Text == "TEMPLATE" {
				payload := &Payload{
					TemplateType: "generic",
					Elements: []Element{
						{
							Title:    "dennougorilla",
							ImageURL: "https://user-images.githubusercontent.com/28649418/45468977-6260f700-b762-11e8-80c3-15fd19c8aa5f.jpeg",
							Subtitle: "Where Do We Come From? What Are We? Where Are We Going?",
							Buttons: []Button{
								Button{
									Type:  "web_url",
									URL:   "https://dennougorilla.tk",
									Title: "View Website",
								},
							},
							DefaultAction: &DefaultAction{
								Type:                "web_url",
								URL:                 "https://github.com/dennougorilla",
								MessengerExtensions: false,
								WebViewHeightRatio:  "tall",
							},
						},
					},
				}
				tmp := template.NewTemplate(senderID, payload)
				PostAction(tmp)
			} else if event.Message.Text == "LIST-TEMPLATE" {
				payload := &Payload{
					TemplateType:    "list",
					TopElementStyle: "compact",
					Elements: []Element{
						{
							Title:    "Hello-1",
							ImageURL: "https://avatars0.githubusercontent.com/u/13715034?s=460&v=4",
							Subtitle: "Hello!!",
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   "https://github.com/uzimaru0000",
									Title: "View Website",
								},
							},
							DefaultAction: &DefaultAction{
								Type:                "web_url",
								URL:                 "https://github.com/uzimaru0000",
								MessengerExtensions: false,
								WebViewHeightRatio:  "tall",
							},
						},
						{
							Title:    "Hello-2",
							ImageURL: "https://avatars0.githubusercontent.com/u/13715034?s=460&v=4",
							Subtitle: "World!",
							Buttons: []Button{
								{
									Type:  "web_url",
									URL:   "https://github.com/uzimaru0000",
									Title: "View Website",
								},
							},
							DefaultAction: &DefaultAction{
								Type:                "web_url",
								URL:                 "https://github.com/uzimaru0000",
								MessengerExtensions: false,
								WebViewHeightRatio:  "tall",
							},
						},
					},
				}

				tmp := template.NewTemplate(senderID, payload)
				PostAction(tmp)
			} else if event.Message.Attachments != nil {
				if &event.Message.Attachments[0].Payload.Coordinates != nil {
					sendTextMessage(senderID, strconv.FormatFloat(event.Message.Attachments[0].Payload.Coordinates.Lat, 'f', 6, 64)+","+strconv.FormatFloat(event.Message.Attachments[0].Payload.Coordinates.Long, 'f', 6, 64))
				}
			} else if event.Message.QuickReply != nil && event.Message.QuickReply.Payload != "" {
				switch event.Message.QuickReply.Payload {
				case "a":
					sendTextMessage(senderID, "You selected a")

				case "b":
					sendTextMessage(senderID, "You selected b")

				case "c":
					sendTextMessage(senderID, "You selected c")
				default:
					sendTextMessage(senderID, "Payload: "+event.Message.QuickReply.Payload)
				}
			}
		}
	}
	fmt.Fprintf(w, "Success")
}

func sendQuickReplies(senderID string, text string, quickReplies []QuickReplies) {
	recipient := new(Recipient)
	recipient.ID = senderID
	m := new(SendMessage)
	m.Recipient = recipient
	m.Message = &SendingMessage{QuickReplies: quickReplies}
	m.Message.Text = text
	PostAction(m)
}

func sendTextMessage(senderID string, text string) {
	recipient := new(Recipient)
	recipient.ID = senderID
	m := new(SendMessage)
	m.Recipient = recipient
	m.Message = &SendingMessage{Text: text}
	PostAction(m)
}

func PostAction(m *SendMessage) {
	b, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
	}

	req, err := http.NewRequest("POST", EndPoint, bytes.NewBuffer(b))
	if err != nil {
		log.Print(err)
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Duration(30 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}

	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Print(err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}
