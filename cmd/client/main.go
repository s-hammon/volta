package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/s-hammon/volta/internal/entity"
)

var (
	cursorID  int = 0
	voltaURL  string
	makatoURL string
	client    *http.Client
)

const (
	getProcedureSpecialtyEndpoint string = "procedure/specialty"
	postPredictEndpoint           string = "predict"
	putProcedureEndpoint          string = "procedure"
)

func init() {
	voltaURL = os.Getenv("VOLTA_URL")
	makatoURL = os.Getenv("MAKATO_URL")
}

type errResponse struct {
	Message string `json:"message"`
}

func main() {
	client = &http.Client{Timeout: time.Duration(5 * time.Second)}
	procURL, err := url.JoinPath(voltaURL, getProcedureSpecialtyEndpoint)
	if err != nil {
		log.Fatalf("couldn't create request URL: %v", err)
	}
	req := newGetProceduresForSpecialtyRequest(procURL, cursorID)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("error making get procedure request: %v", err)
	}
	if res == nil {
		log.Fatal("request to volta did not return a payload!")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("error closing request body: %v", err)
		}
	}()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("error reading payload: %v", err)
	}
	switch res.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError:
		var errResp errResponse
		if err = json.Unmarshal(data, &errResp); err != nil {
			log.Fatalf("error unmarshaling error: %v", err)
		}
		log.Fatalf("request failed: %s", errResp.Message)
	case http.StatusNotFound:
		log.Println("no procedures to update. exiting...")
		os.Exit(0)
	case http.StatusOK:
		var procedures []entity.Procedure
		if err = json.Unmarshal(data, &procedures); err != nil {
			log.Fatalf("error unmarshaling procedures: %v", err)
		}
		predURL, err := url.JoinPath(makatoURL, postPredictEndpoint)
		if err != nil {
			log.Fatalf("couldn't create predict URL: %v", err)
		}
		predictions, err := getPredictions(newPostPredictRequest(predURL, procedures))
		if err != nil {
			log.Fatalf("couldn't get predictions: %v", err)
		}
		type result struct {
			ID        int    `json:"id"`
			Specialty string `json:"specialty"`
		}
		resultRequest := make([]result, len(predictions))
		for i, prediction := range predictions {
			resultRequest = append(resultRequest, result{
				ID:        procedures[i].ID,
				Specialty: prediction,
			})
		}
		finalPayload, err := json.Marshal(resultRequest)
		if err != nil {
			log.Fatalf("error marshaling final payload: %v", err)
		}
		putURL, err := url.JoinPath(voltaURL, putProcedureEndpoint)
		if err != nil {
			log.Fatalf("couldn't create final request URL: %v", err)
		}
		req, err := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(finalPayload))
		if err != nil {
			log.Fatalf("error making final volta request: %v", err)
		}
		_, err = client.Do(req)
		if err != nil {
			log.Fatalf("coulnd't execute final request: %v", err)
		}
		// TODO: finish the rest tomorrow
	}
}

func getPredictions(r *http.Request) ([]string, error) {
	res, err := client.Do(r)
	if err != nil {
		log.Fatalf("error making makato request: %v", err)
	}
	if res.Body == nil {
		log.Fatal("request to makato did not return a payload!")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("error closing request body: %v", err)
		}
	}()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		var errStruct struct {
			Error string `json:"error"`
		}
		if err = json.Unmarshal(data, &errStruct); err != nil {
			return nil, err
		}
		return nil, errors.New(errStruct.Error)
	}

	var result struct {
		Prediction []string `json:"prediction"`
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Prediction, nil
}

var newPostPredictRequest = func(predURL string, procedures []entity.Procedure) *http.Request {
	makatoRequest := struct {
		Input []string `json:"input"`
	}{make([]string, len(procedures))}
	for i, procedure := range procedures {
		makatoRequest.Input[i] = procedure.Description
	}
	data, err := json.Marshal(makatoRequest)
	if err != nil {
		panic(err) // it...should work...
	}
	req, err := http.NewRequest(http.MethodPost, predURL, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	return req
}

var newGetProceduresForSpecialtyRequest = func(procURL string, cursorID int) *http.Request {
	path, err := url.Parse(procURL)
	if err != nil {
		panic(err) // it should work...
	}
	q := url.Values{}
	q.Add("cursor_id", strconv.Itoa(cursorID))
	path.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, path.String(), nil)
	if err != nil {
		panic(err) // it should also work...
	}
	return req
}
