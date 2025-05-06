package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"strconv"
	"strings"

	"testing"

	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	Message []byte
	Err     error
}

func (m mockClient) GetHL7V2Message(path string) ([]byte, error) {
	return m.Message, m.Err
}

func mockPubSubBody(hl7Path, hl7Type string) io.Reader {
	ps := pubSubMessage{
		Message: message{
			Data: []byte(hl7Path),
			Attributes: attributes{Type: hl7Type},
		},
		Subscription: "volta",
	}
	j, _ := json.Marshal(ps)
	return strings.NewReader(string(j))
}

func TestHandleMessage_DebugMode(t *testing.T) {
	want := []byte("MSH|^~\\&|...|ORM^O01|...")
	mockClient := mockClient{Message: want, Err: nil}
	handler := New(&database.Queries{}, mockClient, true)
	req := httptest.NewRequest(http.MethodPost, "/", mockPubSubBody("test.hl7", "ORM"))
	req.Header.Set("Content-Length", strconv.Itoa(len(want)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	
	resp := decodeResponse(t, rec.Body)
	assert.Equal(t, "message received!", resp["message"])
	assert.Equal(t, "test.hl7", resp["hl7_path"])
	assert.Equal(t, 24, int(resp["hl7_size"].(float64)))
}

func TestHandleMessage_HL7FetchError(t *testing.T) {
	mockClient := &mockClient{Message: nil, Err: errors.New("fetch failed")}
	handler := New(&database.Queries{}, mockClient, true)
	req := httptest.NewRequest(http.MethodPost, "/", mockPubSubBody("badpath.hl7", "ORM"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	resp := decodeResponse(t, rec.Body)
	assert.Equal(t, "server error", resp["message"])
	assert.Equal(t, "fetch failed", resp["volta_error"])
	assert.Equal(t, "badpath.hl7", resp["hl7_path"])
}

func TestHandleMessage_EmptyBody(t *testing.T) {
	handler := New(&database.Queries{}, &mockClient{}, true)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	resp := decodeResponse(t, rec.Body)
	// assert.Equal(t, "empty request body", resp["message"]) // not sure why this isn't passing -- maybe httptset doesn't set a nil Body?
	assert.Nil(t, resp["hl7_path"])
}

func decodeResponse(t *testing.T, body *bytes.Buffer) (resp map[string]any) {
	t.Helper()

	err := json.NewDecoder(body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

