package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

type mockClient struct {
	messages map[string][]byte
}

type mockRecord struct {
	key     string
	msgType string
}

func newMockClient(msg map[string][]byte) *mockClient {
	return &mockClient{
		messages: msg,
	}
}

func (m *mockClient) GetHL7V2Message(messagePath string) (hl7.Message, error) {
	raw, ok := m.messages[messagePath]
	if !ok {
		return nil, errors.New("message not found")
	}
	msg, err := hl7.NewMessage(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HL7 message: %w", err)
	}

	return msg, nil
}

type mockRepo struct {
	mu   sync.Mutex
	orms []models.ORM
	orus []models.ORU
}

func (m *mockRecord) GetMsgType(b []byte) error {
	msgJSON, err := hl7.NewMessage(b)
	if err != nil {
		return err
	}
	msh := &struct {
		MSH struct {
			MsgType map[string]string `json:"MSH.9"`
		} `json:"MSH"`
	}{}

	if err := json.Unmarshal(msgJSON, msh); err != nil {
		return err
	}
	msgType := msh.MSH.MsgType["MSH.9.1"]
	if msgType == "" {
		return errors.New("message type not found")
	}
	m.msgType = strings.TrimSpace(msgType)
	return nil
}

func NewMockRepo() *mockRepo {
	return &mockRepo{
		mu:   sync.Mutex{},
		orms: []models.ORM{},
		orus: []models.ORU{},
	}
}

func (m *mockRepo) UpsertORM(ctx context.Context, orm models.ORM) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if orm.MSH.Type.Type != "ORM" {
		return fmt.Errorf("invalid message type: %s", orm.MSH.Type.Type)
	}
	m.orms = append(m.orms, orm)
	return nil
}

func (m *mockRepo) InsertORU(ctx context.Context, oru models.ORU) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oru.MSH.Type.Type != "ORU" {
		return fmt.Errorf("invalid message type: %s", oru.MSH.Type.Type)
	}
	m.orus = append(m.orus, oru)
	return nil
}

func TestHandleMessage(t *testing.T) {
	entries, err := hl7.HL7.ReadDir("test_hl7")
	if err != nil {
		t.Fatalf("failed to read embedded test directory: %v", err)
	}

	msg := make(map[string][]byte)
	records := []mockRecord{}
	for _, entry := range entries {
		name := entry.Name()
		data, err := hl7.HL7.ReadFile(filepath.Join("test_hl7", name))
		if err != nil {
			t.Fatalf("failed to read test file %s: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("file %s is empty", name)
		}
		msg[name] = data

		record := mockRecord{key: name}
		if err := record.GetMsgType(data); err != nil {
			t.Fatalf("failed to get message type from file %s: %v", name, err)
		}
		records = append(records, record)
	}
	client := newMockClient(msg)

	repo := NewMockRepo()
	// for each key in encodedKeys, run a test in parallel
	for i, record := range records {
		t.Run(fmt.Sprintf("message-%d", i), func(t *testing.T) {
			psMessage := &pubSubMessage{
				Message: message{
					Data:       []byte(record.key),
					Attributes: attributes{Type: record.msgType},
				},
			}
			data, err := json.Marshal(psMessage)
			if err != nil {
				t.Fatalf("failed to marshal message: %v", err)
			}
			api := New(repo, client, false)

			req := newPostMsgRequest(data)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				if psMessage.Message.Attributes.Type != "ADT" {
					t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
				} else {
					if w.Code != http.StatusNotImplemented {
						t.Errorf("got status %d, want %d", w.Code, http.StatusNotImplemented)
					}
				}
			}
		})
	}
}

func newPostMsgRequest(data []byte) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	return req
}
