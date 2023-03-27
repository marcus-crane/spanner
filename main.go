package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/vmihailenco/msgpack/v5"
)

// Span represents information about a single HTTP request.
// It reflects a single unit of work and metadata about that
// work unit such as how long it took, whether it succeeded
// and other useful metrics and tags to report.
type Span struct {
	Service  string             `msgpack:"service"`
	Name     string             `msgpack:"name"`
	Resource string             `msgpack:"resource"`
	TraceID  uint64             `msgpack:"trace_id"`
	SpanID   uint64             `msgpack:"span_id"`
	ParentID uint64             `msgpack:"parent_id"`
	Start    int64              `msgpack:"start"`
	Duration int64              `msgpack:"duration"`
	Error    int32              `msgpack:"error"`
	Meta     map[string]string  `msgpack:"meta"`
	Metrics  map[string]float64 `msgpack:"metrics"`
	Type     string             `msgpack:"types"`
}

// Trace is a group of spans that share the same trace ID.
// They may reflect multiple units of work within a single
// service or units of work across multiple services
type Trace []Span

// Traces are what we receive from the APM agents. Each
// APM agent will bundle up multiple traces together
// and send them off to an instance of the Datadog agent.
// This happens around every 10 seconds or so.
type Traces []Trace

func main() {
	verbose := flag.Bool("verbose", false, "verbose mode outputs all information available for each span")
	nameFilter := flag.String("name", "", "filter spans by name (must include the given string value). Not usable with verbose mode.")
	flag.Parse()

	traceHandler := func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		if len(body) == 1 {
			log.Printf("Received initial connection from %s on %s", r.UserAgent(), r.RemoteAddr)
			return
		}
		var traces Traces
		err = msgpack.Unmarshal(body, &traces)
		if err != nil {
			log.Println("Failed to unpack traces")
		}
		if *verbose {
			log.Println(spew.Sdump(traces))
		} else {
			for _, trace := range traces {
				for _, span := range trace {
					if *nameFilter == "" || strings.Contains(span.Name, *nameFilter) {
						log.Printf("%s\n\033[34m%s\033[0m\n", span.Name, span.Resource)
					}
				}
			}
		}
		fmt.Fprintf(w, "OK")
	}

	http.HandleFunc("/", traceHandler)
	log.Print("Spanner is listening on 127.0.0.1:8126")
	log.Fatal(http.ListenAndServe(":8126", nil))
}
