package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

const address = "SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q"

type addressCriteria struct {
	Address string
	Field   string
	Want    string
}

func TestStreamAddresses(t *testing.T) {
	criteria := []addressCriteria{
		{"SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q", "ok", "address registered"},
		{"1FwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q", "error", "address not a STREAM address"},
		{"SFwExaKH1iuZiK9gW3W2dnRQZewcmGkv6q", "error", "address format is invalid"}}

	os.RemoveAll("SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q")
	for _, c := range criteria {
		data := map[string]string{"address": c.Address}
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(data)
		resp, err := http.Post("http://localhost:8080/stream", "application/json", buffer)
		if err != nil {
			t.Error("error in POSTING", err)
		}
		// decode response
		var v map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			t.Error("error in json decoding body", err)
		}

		if v[c.Field] != c.Want {
			t.Errorf("Expected [%s], got [%s]", c.Want, v[c.Field])
		}
	}
}

func TestMissingFieldAddress(t *testing.T) {
	data := map[string]string{"addressy": address}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(data)
	resp, err := http.Post("http://localhost:8080/stream", "application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}
	// decode response
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Error("error in json decoding body", err)
	}

	if v["error"] != "missing needed field, address" {
		t.Error("Expected [missing needed field, address], got", v["error"])
	}
}

func TestStreamBogusJSON(t *testing.T) {
	buffer := bytes.NewBufferString("\"address\": \"" + address + "\"") // not a json object
	resp, err := http.Post("http://localhost:8080/stream", "application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}
	// decode response
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Error("error in json decoding body", err)
	}

	if !strings.Contains(v["error"].(string), "unable to parse JSON") {
		t.Error("Expected [unable to parse JSON...], got", v["error"])
	}
}

func TestStreamExisiting(t *testing.T) {
	os.RemoveAll(address)
	data := map[string]string{"address": address}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(data)
	fmt.Print(buffer)
	resp, err := http.Post("http://localhost:8080/stream", "application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}

	// post again
	data = map[string]string{"address": address}
	buffer = new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(data)
	fmt.Print(buffer)
	resp, err = http.Post("http://localhost:8080/stream", "application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}

	// decode response
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Error("error in json decoding body", err)
	}

	if !strings.Contains(v["error"].(string), "unable to create address") {
		t.Error("Expected [unable to create address...], got", v["error"])
	}
}

func TestStreamMessage(t *testing.T) {
	os.RemoveAll(address)
	data := map[string]string{"address": address}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(data)
	resp, err := http.Post("http://localhost:8080/stream", "application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}

	// a message
	buffer = bytes.NewBufferString("æ a utf-8 message ʩ")
	resp, err = http.Post("http://localhost:8080/stream/"+address+"/message",
		"application/json", buffer)
	if err != nil {
		t.Error("error in POSTING", err)
	}

	// decode response
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Error("error in json decoding body", err)
	}

	// TODO: see if it is a valid time-stamp
	_, ok := v["ok"].(string)
	if !ok {
		t.Error("Expected an ok response")
	}

	// see if there is a location header
	if !strings.Contains(resp.Header.Get("Location"),
		"/stream/SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q/message/") {
		t.Error("Location header not set")
	}
}
