package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

var validORM = hl7.Message{
	"MSH": map[string]interface{}{
		"MSH.9": map[string]interface{}{
			"1": "ORM",
			"2": "O01",
		},
	},
}
var vORMEncKey = base64.StdEncoding.EncodeToString([]byte("validORM"))

var validORU = hl7.Message{
	"MSH": map[string]interface{}{
		"MSH.9": map[string]interface{}{
			"1": "ORU",
			"2": "R01",
		},
	},
}
var vORUEncKey = base64.StdEncoding.EncodeToString([]byte("validORU"))

var invalidORM = hl7.Message{
	"MSH": map[string]interface{}{
		"MSH.9": map[string]interface{}{
			"1": "invalid",
			"2": "O01",
		},
	},
}
var iORMEncKey = base64.StdEncoding.EncodeToString([]byte("invalidORM"))

var invalidORU = hl7.Message{
	"MSH": map[string]interface{}{
		"MSH.9": map[string]interface{}{
			"1": "invalid",
			"2": "R01",
		},
	},
}
var iORUEncKey = base64.StdEncoding.EncodeToString([]byte("invalidORU"))

type mockClient struct {
	messages map[string]hl7.Message
}

func newMockClient(msg map[string]hl7.Message) *mockClient {
	return &mockClient{
		messages: msg,
	}
}

func (m *mockClient) GetHL7V2Message(messagePath string) (hl7.Message, error) {
	msg, ok := m.messages[messagePath]
	if !ok {
		return nil, errors.New("message not found")
	}

	return msg, nil
}

type mockRepo struct {
	orms []models.ORM
	orus []models.ORU
}

func (m *mockRepo) UpsertORM(ctx context.Context, orm models.ORM) error {
	if orm.MSH.Type.Type != "ORM" {
		return errors.New("invalid message type")
	}
	m.orms = append(m.orms, orm)
	return nil
}

func (m *mockRepo) InsertORU(ctx context.Context, oru models.ORU) error {
	if oru.MSH.Type.Type != "ORU" {
		return errors.New("invalid message type")
	}
	m.orus = append(m.orus, oru)
	return nil
}

func TestHandleMessage(t *testing.T) {
	tests := []struct {
		name string
		data string
		want int
	}{
		{
			name: "inserting ORM",
			data: fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "ORM"}}}`, vORMEncKey),
			want: http.StatusCreated,
		},
		{
			name: "inserting ORU",
			data: fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "ORU"}}}`, vORUEncKey),
			want: http.StatusCreated,
		},
		{
			name: "empty request body",
			data: "",
			want: http.StatusBadRequest,
		},
		{
			name: "invalid message type",
			data: fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "invalid"}}}`, vORMEncKey),
			want: http.StatusBadRequest,
		},
		{
			name: "message not found",
			data: `{"message": {"data": "notFound", "attributes": {"type": "ORM"}}}`,
			want: http.StatusInternalServerError,
		},
		{
			name: "error getting ORM",
			data: fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "ORM"}}}`, iORMEncKey),
			want: http.StatusInternalServerError,
		},
		{
			name: "error getting ORU",
			data: fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "ORU"}}}`, iORUEncKey),
			want: http.StatusInternalServerError,
		},
	}

	repo := &mockRepo{}
	client := newMockClient(map[string]hl7.Message{
		"validORM":   validORM,
		"validORU":   validORU,
		"invalidORM": invalidORM,
		"invalidORU": invalidORU,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New(repo, client)

			req := newPostMsgRequest(tt.data)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.want)
		})
	}
}

func newPostMsgRequest(data string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	return req
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func BenchmarkHandleMessage(b *testing.B) {
	runtime.GOMAXPROCS(1)
	repo := &mockRepo{}
	client := newMockClient(map[string]hl7.Message{"validORM": validORM})

	api := New(repo, client)
	data := fmt.Sprintf(`{"message": {"data": "%s", "attributes": {"type": "ORM"}}}`, vORMEncKey)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := newPostMsgRequest(data)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
	}
}
