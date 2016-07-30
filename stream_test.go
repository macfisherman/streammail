package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

const address = "SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q"

func TestNewStream(t *testing.T) {
	os.RemoveAll(address)
	data := map[string]string{"address": address}
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

	if v["ok"] != "address registered" {
		t.Error("Expected [address registered], got", v["ok"])
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
		t.Error("Expected [missing needed field, address], got", v["ok"])
	}
}
