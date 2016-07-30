package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
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
