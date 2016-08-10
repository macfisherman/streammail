package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "github.com/macfisherman/streammail/streamclient"
)

var uri *string
var address *string

func main() {
    uri = flag.String("uri", "http://localhost:8080/stream/v1", "uri of the stream server")
    address = flag.String("address", "", "address to operate against")
    flag.Parse()

    stream := streamclient.NewStream(*uri, *address)
    command := strings.ToLower(flag.Arg(0))
    switch command {
        case "help":
            help()
        case "list":
            index(stream)
        case "read":
            read(stream, flag.Arg(1))
        case "post":
            post(stream)
        default:
            fmt.Printf("Unknown command %s: valid commands are help, list, read, post", command)
    }
}

func help() {
    fmt.Println("coming soon...")
}

func index(s *streamclient.Stream) {
    list, err := s.GetIndex()
    if err != nil {
        fmt.Print("error:", err)
        return
    }

    if len(list)==0 {
        fmt.Println("No messages")
        return
    }

    for i, msg := range list {
        fmt.Printf("%d: %s\n", i, msg)
    }
}

func read(s *streamclient.Stream, id string) {
    msg, err := s.GetMessage(id)
    if err != nil {
        fmt.Println("error in reading message:", err)
        return
    }

    fmt.Printf("%s\n", msg)
}

func post(s *streamclient.Stream) {
    msg, err := ioutil.ReadAll(os.Stdin)
    if err != nil {
        fmt.Println("error reading stdin:", err)
    }

    s.PostMessage(string(msg))
}
