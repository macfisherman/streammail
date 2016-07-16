package main

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func report_error(w http.ResponseWriter, e string) {
	fmt.Fprintf(w, "{ \"error\": \"%s\" } \n", e)
}

func report_status(w http.ResponseWriter, s string) {
	fmt.Fprintf(w, "{ \"status\": \"%s\" } \n", s)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func Message(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	err := os.Chdir(address)
	if err != nil {
		report_error(w, err.Error())
		return
	}

	filename := time.Now().UTC().Format(time.RFC3339Nano)
	msg, err := os.Create(filename)
	if err != nil {
		report_error(w, err.Error())
		return
	}
	defer msg.Close()

	if _, err := io.Copy(msg, r.Body); err != nil {
		report_error(w, err.Error())
		return
	}

	report_status(w, "saved "+filename)
}

func Register(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	log.Printf("address is: %c:%v", address[0], address)
	if !((address[0] == 'S') || (address[0] == 'R')) {
		report_error(w, "address not a STREAM address")
		return
	}
	_, _, err := base58.CheckDecode(address)
	if err != nil {
		report_error(w, "address format is invalid")
		return
	}

	err = os.Mkdir(ps.ByName("address"), 0755)
	if err != nil {
		report_error(w, "unable to create address:"+err.Error())
	} else {
		report_status(w, "address registered")
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/register/:address", Register)
	router.POST("/post/:address", Message)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
