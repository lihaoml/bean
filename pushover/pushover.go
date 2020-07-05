package pushover

import (
	. "bean"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"net/url"
	"os"
)

const (
	pushoverGlancesURL = "https://api.pushover.net/1/glances.json"
	pushoverMessagesURL = "https://api.pushover.net/1/messages.json"
	statusSuccess = 1
)

type Message struct {
	Title string
	Text string
	Subtext string
	Count int
}

// Response contains the JSON response returned by the pushover.net API
type response struct {
	Status  int      `json:"status"`
	Errors  []string `json:"errors"`
}

func UpdateGlance(apiToken, userKey string, msg Message) error {
	return send(pushoverGlancesURL, apiToken, userKey, msg)
}

func SendMsg(apiToken, userKey string, msg Message) error {
	return send(pushoverMessagesURL, apiToken, userKey, msg)
}

func send(pourl, apiToken, userKey string, msg Message) (err error) {
	godotenv.Overload(BeanexAccountPath() + "tgbot.env")
	if os.Getenv(apiToken) == "" {
		panic("Error loading pushover api token")
	}
	if os.Getenv(userKey) == "" {
		panic("Error loading pushover user key")
	}
	// Initalise an empty Response
	r := &response{}
	m := url.Values{}
	m.Set("token", os.Getenv(apiToken))
	m.Set("user", os.Getenv(userKey))
	m.Set("title", msg.Title)
	m.Set("text", msg.Text)
	m.Set("subtext", msg.Subtext)
	m.Set("count", fmt.Sprint(msg.Count))
	// Send the message the the pushover.net API
	resp, err := http.PostForm(pourl, m)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Decode the json response from pushover.net in to our Response struct
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return err
	}
	// Check to see if pushover.net set the status to indicate an error without providing and explanation
	if r.Status != statusSuccess {
		if len(r.Errors) < 1 {
			return errors.New("Recieved a status code indicating an error but did not receive an error message from pushover.net")
		}
		return errors.New(r.Errors[0])
	}
	return  nil
}