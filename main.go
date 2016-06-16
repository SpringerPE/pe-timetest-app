package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    "strconv"
)

type Configuration struct {
    Port                  string
    SlowQueryLogThreshold time.Duration
}

type WrapHTTPHandler struct {
    handler http.Handler
}

type LoggedResponse struct {
    http.ResponseWriter
    status int
}

var config *Configuration

func init() {

    log.SetOutput(os.Stdout)

    config = new(Configuration)
    var success bool

    config.Port, success = os.LookupEnv("PORT")
    if ! success {
        log.Fatal("Please define PORT env variable")
    }
    slow_query_threshold, success := os.LookupEnv("SLOW_QUERY_LOG_THRESHOLD")
    if ! success {
        log.Fatal("Please define SLOW_QUERY_LOG_THRESHOLD env variable")
    }
    slow_query_threshold_int, _ := strconv.Atoi(slow_query_threshold)
    config.SlowQueryLogThreshold = time.Duration(slow_query_threshold_int)


}


func (wrappedHandler *WrapHTTPHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
    loggedWriter := &LoggedResponse{ResponseWriter: writer, status: 200}
    start := time.Now()
    wrappedHandler.handler.ServeHTTP(loggedWriter, request)
    elapsed := time.Since(start)
    if elapsed > config.SlowQueryLogThreshold {
      log.Printf("[RemoteAddr: %s] URL: %s, STATUS: %d, Time Elapsed: %d ns.\n",
        request.RemoteAddr, request.URL, loggedWriter.status, elapsed)
  }
}

func (loggedResponse *LoggedResponse) WriteHeader(status int) {
    loggedResponse.status = status
    loggedResponse.ResponseWriter.WriteHeader(status)
}

func hello(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintln(w, "hello, world!")
}

func main() {

    http.HandleFunc("/", hello)
    err := http.ListenAndServe(":"+ config.Port, &WrapHTTPHandler{http.DefaultServeMux})
    if err != nil {
        log.Printf("ListenAndServe:", err)
    }
}

