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
const baseURI = "http://localhost:8080/stream"

type addressCriteria struct {
	Address string
	Field   string
	Want    string
}

func newStream(t *testing.T, address string) *http.Response {
	return postMap(t, map[string]string{"address": address}, baseURI)
}

func postString(t *testing.T, data string, uri string) *http.Response {
	buffer := bytes.NewBufferString(data)
	resp, err := http.Post(uri, "application/json", buffer)
	if err != nil {
		t.Fatal("fatal error in posting:", err)
	}

	return resp
}

func postMapAsJSON(t *testing.T, d map[string]string, uri string) *http.Response {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(d)
	resp, err := http.Post(uri, "application/json", buffer)
	if err != nil {
		t.Fatal("fatal error in posting:", err)
	}

	return resp
}

func decodeResponse(t *testing.T, r *http.Response) map[string]interface{} {
	var v map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		t.Fatal("fatal error in decoding response:", err)
	}

	return v
}

func TestStreamAddresses(t *testing.T) {
	criteria := []addressCriteria{
		{"SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q", "ok", "address registered"},
		{"1FwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q", "error", "address not a STREAM address"},
		{"SFwExaKH1iuZiK9gW3W2dnRQZewcmGkv6q", "error", "address format is invalid"}}

	os.RemoveAll("SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q")
	for _, c := range criteria {
		resp := newStream(t, c.Address)
		v := decodeResponse(t, resp)

		if v[c.Field] != c.Want {
			t.Errorf("Expected [%s], got [%s]", c.Want, v[c.Field])
		}
	}
}

func TestMissingFieldAddress(t *testing.T) {
	resp := postMapAsJSON(t, map[string]string{"addressy": address}, baseURI)
	v := decodeResponse(t, resp)
	if v["error"] != "missing needed field, address" {
		t.Error("Expected [missing needed field, address], got", v["error"])
	}
}

func TestStreamBogusJSON(t *testing.T) {
	resp := postString(t, "\"address\": \""+address+"\"", baseURI)
	v := decodeResponse(t, resp)
	if !strings.Contains(v["error"].(string), "unable to parse JSON") {
		t.Error("Expected [unable to parse JSON...], got", v["error"])
	}
}

func TestStreamExisting(t *testing.T) {
	os.RemoveAll(address)
	_ = newStream(t, address)
	// post again
	resp := newStream(t, address)
	v := decodeResponse(t, resp)
	if !strings.Contains(v["error"].(string), "unable to create address") {
		t.Error("Expected [unable to create address...], got", v["error"])
	}
}

func TestStreamMessage(t *testing.T) {
	os.RemoveAll(address)
	_ = newStream(t, address)

	// a message
	resp := postString(t, "æ a utf-8 message ʩ", baseURI+"/"+address+"/message")
	v := decodeResponse(t, resp)
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
