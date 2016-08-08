package streamclient

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

const address = "SFwExaKH1iu2iK9gW3W2dnRQZewcmGkv6q"
const baseURI = "http://localhost:8080/stream/v1"
const baseDir = "../" + address

func cleanup() {
	os.RemoveAll(baseDir)
}

func TestRegisterStream(t *testing.T) {
	cleanup()
	
	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}
}

func TestStreamExisting(t *testing.T) {
	cleanup()
	
	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}
	
	if err := stream.Register(); !strings.Contains(err.Error(), "unable to create address") {
		t.Error("error registering stream", err)
	}
}

func TestStreamMessage(t *testing.T) {
	cleanup()

	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}

	// a message
	if err:= stream.PostMessage("æ a utf-8 message ʩ"); err != nil {
		t.Error("error posting message. Got", err)
	}
}

func TestStreamIndex(t *testing.T) {
	cleanup()
	
	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}

	// a bunch of messages
	for i := 0; i < 10; i++ {
		_ = stream.PostMessage("message "+strconv.Itoa(i))
	}
	
	list, err := stream.GetIndex()
	if err != nil {
		t.Fatal("error getting index", err)
	}
	
	l := len(list)
	if l != 10 {
		t.Error("Expected 10 items, got", l)
	}
}

func TestStreamIndexFrom(t *testing.T) {
	cleanup()

	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}

	// a bunch of messages
	for i := 0; i < 120; i++ {
		_ = stream.PostMessage("message "+strconv.Itoa(i))
	}
	
	list, err := stream.GetIndex()
	if err != nil {
		t.Fatal("error getting index", err)
	}
	l := len(list)
	if l != 100 {
		t.Error("Expected 100 items, got", l)
	}

	from := list[99]
	list, err = stream.GetIndexFrom(from, 4)
	l = len(list)
	if l != 4 {
		t.Error("Expected 4 items, got", l)
	}

	from = list[3]
	list, err = stream.GetIndexFrom(from, 0)
	l = len(list)
	if l != 18 {
		t.Error("Expected 18 items, got", l)
	}
}

func TestStreamGetMessage(t *testing.T) {
	cleanup()

	stream := NewStream(baseURI, address)
	if err := stream.Register(); err != nil {
		t.Fatal("error registering stream", err)
	}

	if err := stream.PostMessage("message one"); err != nil {
		t.Fatal("error posting message", err)
	}
	
	// get first message
	list, err := stream.GetIndex()
	if err != nil {
		t.Fatal("error getting index", err)
	}
	
	msg, err := stream.GetMessage(list[0])
	if err != nil {
		t.Fatal("error getting body", err)
	}
	
	if string(msg) != "message one" {
		t.Error("expected [message one], got", msg)
	}
}