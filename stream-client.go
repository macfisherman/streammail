package StreamClient

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

type Stream struct {
	BaseURI string
	Address string
}

func decodeResponse(r *http.Response) (map[string]interface{}, error) {
	var v map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func decodeResponseArray(r *http.Response) ([]interface{}, error) {
	var v []interface{}

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}


func postMap(d map[string]string, uri string) (*http.Response, error) {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(d)
	return http.Post(uri, "application/json", buffer)
}

func postString(data string, uri string) (*http.Response, error) {
	buffer := bytes.NewBufferString(data)
	return http.Post(uri, "application/json", buffer)
}

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

func NewStream(uri string, address string) *Stream {
	return &Stream{ BaseURI: uri, Address: address }
}

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

func (s *Stream) GetIndex() ([]string, error) {
	uri := s.BaseURI + "/" + s.Address
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
