// Copyright 2016 Jeff Macdonald <macfisherman@gmail.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A Go SDK that implements the Stream API
package streamclient

import (
	"strconv"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"errors"
//	"fmt"
)

const VERSION = "v1"
const API = "stream"
const APP = "/"+API+"/"+VERSION

// The Stream to do actions on.
type Stream struct {
	BaseURI string
	Address string
}

// Helper function to decode a http.Response body that
// contains a JSON object into a native go map.
func decodeResponse(r *http.Response) (map[string]interface{}, error) {
	var v map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Helper function to decode a http.Response body that
// contains JSON array into native go array.
func decodeResponseArray(r *http.Response) ([]interface{}, error) {
	var v []interface{}

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Helper function to POST a map to an HTTP endpoint.
func postMap(d map[string]string, uri string) (*http.Response, error) {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(d)
	return http.Post(uri, "application/json", buffer)
}

// Helper function to POST a string to an HTTP endpoint.
func postString(data string, uri string) (*http.Response, error) {
	buffer := bytes.NewBufferString(data)
	return http.Post(uri, "application/json", buffer)
}

// Helper function to GET an HTTP endpoint.
// Sets the following headers:
// Content-Type: application/stream+json
// Accept: application/json
func get(uri string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/stream+json") // vnd.api should be something stream specific?
	req.Header.Set("Accept", "application/json")
	return client.Do(req)
}

// Create a Stream object.
// uri must be in the form https://host/stream/v1
// Address must be in base58check format.
// See https://github.com/macfisherman/streammail/Stream-Address
// for more information about Stream addresses.
// All methods of the resulting object use the Stream address
// provided here.
func NewStream(uri string, address string) *Stream {
	return &Stream{ BaseURI: uri, Address: address }
}

// Register the Stream address (which was specified with NewStream)
// with the server.
// This only has to be done once with a server.
func (s *Stream) Register() error {
	resp, err := postMap(map[string]string{"address": s.Address}, s.BaseURI)
	if err != nil {
		return err
	}
	
	m, err := decodeResponse(resp)
	if err != nil {
		return err
	}
	
	_, ok := m["error"].(string)
	if ok {
		return errors.New(m["error"].(string))
	}
	
	_, ok = m["ok"].(string)
	if !ok {
		return errors.New("server did not return ok or an error")
	}
	
	return nil
}

// Post a message to the server.
func (s *Stream) PostMessage(message string) error {
	resp, err := postString(message, s.BaseURI+"/"+s.Address+"/message")
	if err != nil {
		return err
	}
	
	m, err := decodeResponse(resp)
	if err != nil {
		return err
	}
	
	if m["ok"] == "" {
		return errors.New("server did not return ok")
	}
	
	return nil
}

// Get a message 'id' from server.
func (s *Stream) GetMessage(id string) ([]byte, error) {
	resp, err := get(s.BaseURI+"/"+s.Address+"/message/"+id)
	if err != nil {
		return nil, err
	}
	
	defer resp.Body.Close()
	message, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return message, nil
}

// Get an array of message 'ids', starting with message id 'from'
// and up to 'count' ids.
func (s *Stream) GetIndexFrom(from string, count int) ([]string, error) {
	uri := s.BaseURI + "/" + s.Address + "/index?from=" + from
	if count > 0 {
		uri = uri + "&count=" + strconv.Itoa(count)
	}
	
	resp, err := get(uri)
	if err != nil {
		return nil, err
	}
	
	a, err := decodeResponseArray(resp)
	if err != nil {
		return nil, err
	}
	
	list := make([]string, len(a))
	for i := range(list) {
		list[i] = a[i].(string)
	}
	
	return list, nil
}

// Get an array of message 'ids'. Only the first 100
// will be returned. Use GetIndexFrom to retrieve 'ids'
// above 100.
func (s *Stream) GetIndex() ([]string, error) {
	uri := s.BaseURI + "/" + s.Address
	resp, err := get(uri)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode == 404 {
		return nil, errors.New("not found")
	}
	
	a, err := decodeResponseArray(resp)
	if err != nil {
		return nil, err
	}
	
	list := make([]string, len(a))
	for i := range(list) {
		list[i] = a[i].(string)
	}
	
	return list, nil
}
