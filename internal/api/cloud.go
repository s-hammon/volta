package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

type pubSubMessage struct {
	Message struct {
		Data       []byte     `json:"data,omitempty"`
		Attributes attributes `json:"attributes,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type attributes struct {
	Type string `json:"type"`
}

func NewPubSubMessage(body io.Reader) (*pubSubMessage, error) {
	var m pubSubMessage
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	if m.Message.Data == nil {
		return nil, errors.New("empty message data")
	}
	if !slices.Contains([]string{"ORM", "ORU", "ADT"}, m.Message.Attributes.Type) {
		return nil, errors.New("unknown message type")
	}

	return &m, nil
}

type Hl7Client struct {
	httpClient *http.Client
	BaseURL    string
	AuthToken  string
}

func NewHl7Client(baseURL, authToken string, timeout time.Duration) (*Hl7Client, error) {
	return &Hl7Client{
		httpClient: &http.Client{Timeout: timeout},
		BaseURL:    baseURL,
		AuthToken:  authToken,
	}, nil
}

func (h *Hl7Client) GetHL7V2Message(messagePath string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", h.BaseURL, messagePath)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AuthToken))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := h.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var message struct {
		Data []byte `json:"data"`
		Type string `json:"messageType"`
	}

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&message); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	if message.Data == nil {
		return nil, errors.New("empty message data")
	}
	if !slices.Contains([]string{"ORM", "ORU", "ADT"}, message.Type) {
		return nil, errors.New("unknown message type")
	}

	return message.Data, nil
}
