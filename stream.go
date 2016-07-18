package main

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func report_error(w http.ResponseWriter, e string) {
	fmt.Fprintf(w, "{ \"error\": \"%s\" } \n", e)
}

func report_status(w http.ResponseWriter, s string) {
	fmt.Fprintf(w, "{ \"status\": \"%s\" } \n", s)
}

func IndexPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func PostMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	filename := time.Now().UTC().Format(time.RFC3339Nano)
	msg, err := os.Create(address + "/" + filename)
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

func GetMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := ps.ByName("address") + "/" + ps.ByName("id")
	msg, err := os.Open(filepath)
	if err != nil {
		report_error(w, err.Error())
		return
	}
	defer msg.Close()

	if _, err := io.Copy(w, msg); err != nil {
		report_error(w, err.Error())
		return
	}
}

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	vars := r.URL.Query()

	dirHandle, err := os.Open(address)
	if err != nil {
		report_error(w, err.Error())
		return
	}
	defer dirHandle.Close()

	files, err := dirHandle.Readdir(0)
	if err != nil {
		report_error(w, err.Error())
		return
	}

	var names []string
	for _, file := range files {
		if file.Mode().IsRegular() {
			names = append(names, file.Name())
		}
	}

	sort.Strings(names)

	// setup a count - default to 100
	count := 100
	skipTo := vars.Get("from")
	if n := vars.Get("count"); n != "" {
		count, err = strconv.Atoi(n)
		if err != nil {
			report_error(w, err.Error())
			return
		}
	}

	have := 0
	if skipTo != "" {
		var wantedNames []string
		getRemaining := false
		for _, name := range names {
			if name == skipTo {
				getRemaining = true
			}
			if getRemaining {
				wantedNames = append(wantedNames, name)
				have++
				if have == count {
					break
				}
			}
		}

		names = wantedNames
	}

	if count > len(names) {
		count = len(names)
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(names[:count]) // only return count
	if err != nil {
		report_error(w, err.Error())
		return
	}
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
	router.GET("/", IndexPage)
	router.POST("/register/:address", Register)
	router.POST("/post/:address", PostMessage)
	router.GET("/index/:address", Index)
	router.GET("/get/:address/:id", GetMessage)
	router.GET("/hello/:name", Hello)

	n := negroni.Classic()
	n.UseHandler(router)

	log.Fatal(http.ListenAndServe(":8080", n))
}
