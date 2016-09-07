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
	"flag"
)

const VERSION = "v1"
const API = "stream"
const APP = "/"+API+"/"+VERSION

// simple wrapper function to write out golang vars as json
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
	return nil
}

// report an error to client - in JSON
func report_error(w http.ResponseWriter, code int, err string) {
	w.WriteHeader(code)

	WriteJSON(w, map[string]string{
		"error": err,
	})
}

// report a status to a client - in JSON
func report_status(w http.ResponseWriter, code int, v interface{}) error {
	w.WriteHeader(code)

	return WriteJSON(w, v)
}

// very simple index page for now
func IndexPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

// Stream API
// POST /stream/ADDRESS/message
//	The post body contains the message.
//	Adds a message to ADDRESS. Returns a message-id. Messages ids are timestamps in UTC
//	in RFC3339Nano format
//
// This implementation stores each message in a directory <address>, where each
// message is a timestamp. This will allow for simple ordered listings without
// requiring any state from the server.
//
// On success an HTTP 201 with location header is returned.
// On error, an HTTP 409 is returned
//
// NOTE: even though the file is created with nano times, two go routines could come
// up with the same file. Need to fix that.
func PostMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	filename := time.Now().UTC().Format(time.RFC3339Nano)
	path := address + "/" + filename
	msg, err := os.Create(path)
	if err != nil {
		report_error(w, 409, "in creating message file "+path+": "+err.Error())
		return
	}
	defer msg.Close()

	if _, err := io.Copy(msg, r.Body); err != nil {
		report_error(w, 409, "in serializing messagee "+path+": "+err.Error())
		return
	}

	w.Header().Set("Location", "/stream/"+address+"/message/"+filename)
	report_status(w, 201, map[string]string{"ok": filename})
}

// Stream API
// GET /stream/ADDRESS/message/ID
//	gets a single message
//
// On success, returns 200 plus a data blob in the body
// On error, returns either a 404 when the message does no exist
//  or a 409 when unable to return the message due to a system error
func GetMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := ps.ByName("address") + "/" + ps.ByName("id")
	msg, err := os.Open(filepath)
	if err != nil {
		report_error(w, 404, filepath+": "+err.Error())
		return
	}
	defer msg.Close()

	// might want to rethink how msg is just a blob and not a JSON object
	if _, err := io.Copy(w, msg); err != nil {
		report_error(w, 409, err.Error())
		return
	}
}

// Stream API
// GET /stream/ADDRESS
// GET /stream/ADDRESS?count=N
// GET /stream/ADDRESS?from=ID
// GET /stream/ADDRESS?from=ID&count=N
//
//	get message-ids, as a JSON array.
//
// The first form will return up to 100 message-ids starting with the first message.
// The second form will return up to N message-ids, starting with the first message.
// The third form will return up to 100 message-ids starting with message-id ID.
// The forth form will return up to N message-ids starting from message-id ID.
//
// In all cases, message-ids are returned in increasing chronilogical order.
//
// The On success, returns a JSON array (up to N or 100 elements) of message-ids
// On error, returns either
//   404 if the address does not exist or
//   409 if the server has a problem reading the directory where the messages are
//   400 if the count N is not a number
//   409 if the server cannot encode the data as JSON
func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	vars := r.URL.Query()

	dirHandle, err := os.Open(address)
	if err != nil {
		report_error(w, 404, err.Error())
		return
	}
	defer dirHandle.Close()

	files, err := dirHandle.Readdir(0)
	if err != nil {
		report_error(w, 409, err.Error())
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
			report_error(w, 400, "invalid number "+n+" :"+err.Error())
			return
		}
	}

	// advance to message-id specified in parameter from
	// and collect that message-id and the remaining message-ids
	// up to count
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
		report_error(w, 409, err.Error())
		return
	}
}

// Stream API
// POST /stream
// with JSON:	{ "address": ADDRESS }
//	Register address with server
//	ADDRESS MUST conform to base58Check
//
// On success returns an HTTP 201 with a Location header
// On error returns either:
//  400 - unable to parse input JSON
//  400 - missing JSON field
//  400 - invalid address
//  400 - invalid Stream address (must start with an S or R)
//  409 - unable to create the Stream address directory
//
func Register(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var fields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		r.Body.Close()
		report_error(w, 400, "unable to parse JSON: "+err.Error())
		return
	}

	address, ok := fields["address"].(string)
	if !ok {
		report_error(w, 400, "missing needed field, address")
		return
	}

	log.Printf("address is: %v", address)
	if !((address[0] == 'S') || (address[0] == 'R')) {
		report_error(w, 400, "address not a STREAM address")
		return
	}
	if _, _, err := base58.CheckDecode(address); err != nil {
		report_error(w, 400, "address format is invalid")
		return
	}

	// first go routine gets to create address, others
	// will get OS error.
	if err := os.Mkdir(address, 0755); err != nil {
		report_error(w, 409, "unable to create address:"+err.Error())
	} else {
		w.Header().Set("Location", "/stream/"+address)
		report_status(w, 201, map[string]string{"ok": "address registered"})
	}
}

func main() {
	use_tls := flag.Bool("tls", true, "enable/disable tls")
	flag.Parse()

	router := httprouter.New()
	router.GET("/", IndexPage)
	router.POST(APP, Register)
	router.POST(APP+"/:address/message", PostMessage)
	router.GET(APP+"/:address", Index)
	router.GET(APP+"/:address/index", Index)
	router.GET(APP+"/:address/message/:id", GetMessage)

	n := negroni.Classic()
	n.UseHandler(router)

	if *use_tls {
		log.Print("serving with TLS")
		log.Fatal(http.ListenAndServeTLS(":8080", "server.pem", "server.key", n))
	} else {
		log.Print("serving without TLS")
		log.Fatal(http.ListenAndServe(":8080", n))
	}
}
