package main

import (
	"bytes"
	"encoding/json"
//	"fmt"
	"net/http"
	"os"
	"strconv"
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

func getIndexFrom(t *testing.T, address string, from string, count int) *http.Response {
	uri := baseURI + "/" + address + "/index?from=" + from
	if count > 0 {
		uri = uri + "&count=" + strconv.Itoa(count)
	}
	return (get(t, uri))
}

func getIndex(t *testing.T, address string) *http.Response {
	return (get(t, baseURI+"/"+address))
}

func postMessage(t *testing.T, address string, message string) *http.Response {
	return postString(t, message, baseURI+"/"+address+"/message")
}

func newStream(t *testing.T, address string) *http.Response {
	return postMap(t, map[string]string{"address": address}, baseURI)
}

func get(t *testing.T, uri string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal("error creating request object:", err)
	}

	req.Header.Set("Content-Type", "application/vnd.api+json") // vnd.api should be something stream specific?
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error with GET request", err)
	}

	return resp
}

func postString(t *testing.T, data string, uri string) *http.Response {
	buffer := bytes.NewBufferString(data)
	resp, err := http.Post(uri, "application/json", buffer)
	if err != nil {
		t.Fatal("fatal error in posting:", err)
	}

	return resp
}

func postMap(t *testing.T, d map[string]string, uri string) *http.Response {
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

func decodeResponseArray(t *testing.T, r *http.Response) []interface{} {
	var v []interface{}

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
	resp := postMap(t, map[string]string{"addressy": address}, baseURI)
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
	resp := postMessage(t, address, "æ a utf-8 message ʩ")
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

func TestStreamIndex(t *testing.T) {
	os.RemoveAll(address)
	_ = newStream(t, address)

	// a bunch of messages
	for i := 0; i < 10; i++ {
		_ = postMessage(t, address, "message "+strconv.Itoa(i))
	}
	resp := getIndex(t, address)
	v := decodeResponseArray(t, resp)
	l := len(v)
	if l != 10 {
		t.Error("Expected 10 items, got", l)
	}
}

func TestStreamIndexFrom(t *testing.T) {
	os.RemoveAll(address)
	_ = newStream(t, address)

	// a bunch of messages
	for i := 0; i < 120; i++ {
		_ = postMessage(t, address, "message "+strconv.Itoa(i))
	}
	resp := getIndex(t, address)
	v := decodeResponseArray(t, resp)
	l := len(v)
	if l != 100 {
		t.Error("Expected 100 items, got", l)
	}

	from := v[99].(string)
	resp = getIndexFrom(t, address, from, 4)
	v = decodeResponseArray(t, resp)
	l = len(v)
	if l != 4 {
		t.Error("Expected 4 items, got", l)
	}

	from = v[3].(string)
	resp = getIndexFrom(t, address, from, 0)
	v = decodeResponseArray(t, resp)
	l = len(v)
	if l != 16 {
		t.Error("Expected 16 items, got", l)
	}
}
