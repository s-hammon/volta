package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5"
	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHL7Store struct {
	saveORMErr error
	saveORUErr error
}

func (m *mockHL7Store) SaveORM(ctx context.Context, order *entity.Order) error {
	return m.saveORMErr
}

func (m *mockHL7Store) SaveORU(ctx context.Context, obs *entity.Observation) error {
	return m.saveORUErr
}

func (m *mockHL7Store) GetProcedures(ctx context.Context, cursorID int32) (ret []byte, error error) {
	// mock no records
	if cursorID == 100 {
		return nil, pgx.ErrNoRows
	}
	if cursorID == 11 {
		return nil, fmt.Errorf("internal error")
	}
	procs := []mockProcedureRecord{}
	for i := cursorID; i < cursorID+5; i++ {
		procs = append(procs, mockProcedureRecord{ID: i})
	}
	return json.Marshal(procs)
}

type mockProcedureRecord struct {
	ID int32 `json:"id"`
}

func (m *mockHL7Store) UpdateProcedures(ctx context.Context, data []byte) (int, int, error) {
	var records []mockProcedureRecord
	err := json.Unmarshal(data, &records)
	return len(records), len(records), err
}

type mockHealthcareClient struct {
	message []byte
	err     error
}

func (m *mockHealthcareClient) GetHL7V2Message(path string) ([]byte, error) {
	return m.message, m.err
}

var mockORM = []byte("MSH|^~\\&|SendingApp|SendingFac|ReceivingApp|ReceivingFac|202205271230||ORM^R01|MSGID123|P|2.3|||AL|NE\rPID|1||123456^^^Hospital^MR||Doe^John^A")

func TestHandleMessage_UpdateProcedureSpecialty_Success(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	data := []byte(`[{"id": 1},{"id": 2}]`)
	req := httptest.NewRequest(http.MethodPut, "/procedure", bytes.NewBuffer(data))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NotNil(t, res.Body)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("couldn't close body:", err.Error())
		}
	}()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	got := updateResult{}
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)

	assert.Equal(t, 2, got.RecordsUpdated)
	assert.Equal(t, "62", res.Header.Get("Content-Length"))
}

func TestHandleMessage_GetProcedures_Success(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	req := httptest.NewRequest(http.MethodGet, "/procedure/specialty?cursor_id=1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NotNil(t, res.Body)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("couldn't close body:", err.Error())
		}
	}()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	got := []mockProcedureRecord{}
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)

	assert.Equal(t, 5, len(got))
	for i, g := range got {
		assert.Equal(t, int32(i+1), g.ID)
	}
}

func TestHandleMessage_GetProcedures_InvalidCursorID(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	// no cursor_id provided
	req := httptest.NewRequest(http.MethodGet, "/procedure/specialty", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	require.NotNil(t, res.Body)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("couldn't close body:", err.Error())
		}
	}()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	got := response{}
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)
	require.Equal(t, "must provide value for cursor_id param (int)", got.Message)

	// invalid cursor_id
	req = httptest.NewRequest(http.MethodGet, "/procedure/specialty?cursor_id=one", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res = w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	require.NotNil(t, res.Body)

	respBody, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	got = response{}
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)
}

func TestHandleMessage_GetProcedures_Error(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	// no results
	req := httptest.NewRequest(http.MethodGet, "/procedure/specialty?cursor_id=404", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	// internal error
	req = httptest.NewRequest(http.MethodGet, "/procedure/specialty?cursor_id=11", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res = w.Result()
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandleMessage_ORM_Success(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{message: mockORM}
	handler := New(mockStore, &mockClient, false)

	body, n := newRequestBody(t, "path/to/msg.hl7")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(n))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	assert.NotNil(t, res.Body)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("couldn't close body:", err.Error())
		}
	}()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	got := &response{}
	err = json.Unmarshal(respBody, got)
	require.NoError(t, err)
	assert.Equal(t, response{
		Message:              "message saved",
		RequestContentLength: n,
		HL7Path:              "path/to/msg.hl7",
		HL7Size:              len(mockORM),
		ControlID:            "MSGID123",
	}, *got)
}

func TestHandleMessage_EmptyBody(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandleMEssage_HL7FetchError(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{err: errors.New("fetch error")}
	handler := New(mockStore, &mockClient, false)

	body, n := newRequestBody(t, "path/to/error.hl7")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Length", strconv.Itoa(n))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

}

func TestHandleMessage_BadContentLength(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("some data")))
	req.Header.Set("Content-Length", "pi")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandleMessage_InvalidMethodsPaths(t *testing.T) {
	mockStore := new(mockHL7Store)
	mockClient := mockHealthcareClient{}
	handler := New(mockStore, &mockClient, false)

	for _, method := range []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
	} {
		req := httptest.NewRequest(method, "/", nil)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		res := w.Result()
		require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	}
}

func newRequestBody(t *testing.T, data string) (*bytes.Buffer, int) {
	t.Helper()

	psMsg := &pubSubMessage{
		Message: message{
			Data: []byte(data),
			Attributes: attributes{
				Type: "ORM",
			},
		},
	}
	rBody, err := json.Marshal(psMsg)
	require.NoError(t, err)
	return bytes.NewBuffer(rBody), len(rBody)
}
