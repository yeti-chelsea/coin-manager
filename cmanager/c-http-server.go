package cmanager

import (
	"net/http"
	"strconv"
	"time"
)

type myHandlers func(http.ResponseWriter, *http.Request)

type HttpServer struct {
	// Port Number in which this HTTP server will be listening
	ListenPort int

	// Http server parameters
	ServerRef *http.Server

	// Pointer/reference to the logger interface
	Log_ref *Logger

	// A flag denoting running status
	Running bool

	// Controller for serving http requests
	Controller map[string]myHandlers

	// Channel for sending Request to UDP server
	SendRequestToUdp chan<- string

	// Channel for receiving Response from UDP server
	RespnoseReceiveFromUdp <-chan string
}

func (http_s *HttpServer) LocalRequestHandler(w http.ResponseWriter, r *http.Request) {

	http_s.Log_ref.Debug("Received request for serving locally")

	responseFromUdp := "Unsupported-Query"

	var requestFound = false
	for _,supported_req := range SupportedCurlRequest {
		if supported_req == r.URL.RawQuery {
			requestFound = true
			break
		}
	}

	if ! requestFound {
		w.Write([]byte(responseFromUdp))
		return
	}

	http_s.Log_ref.Debug("Sending Request to UDP server")
	http_s.SendRequestToUdp<- r.URL.RawQuery
	responseFromUdp = <-http_s.RespnoseReceiveFromUdp
	http_s.Log_ref.Debug("Received response from UDP server")

	w.Write([]byte(responseFromUdp))
}

func (http_s *HttpServer) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {

	http_s.Log_ref.Debug("Received request for proxying")
	http_s.Log_ref.Debug(r.URL.Path)
}

func (http_s *HttpServer) AddHandlers(pattern string, handler myHandlers) {
	http_s.Controller[pattern] = handler
}

func (http_s *HttpServer) Init(listenIp string, listenPort int, logRef *Logger) {

	http_s.Log_ref = logRef
	multiplexer := http.NewServeMux()

	http_s.Controller = make(map[string]myHandlers)

	http_s.Log_ref.Debug("Initializing HTTP server")
	http_s.AddHandlers("/rest/lserver", http_s.LocalRequestHandler)
	http_s.AddHandlers("/rest/rproxy", http_s.ProxyRequestHandler)

	for pattern, request_handler := range http_s.Controller {
		http_s.Log_ref.Debug("Registering handler with multiplexer : ", pattern)
		multiplexer.HandleFunc(pattern, request_handler)
	}

	http_s.ServerRef = new(http.Server)
	http_s.ServerRef.Addr = listenIp + ":" + strconv.Itoa(listenPort)
	http_s.ServerRef.ReadTimeout = 5 * time.Second
	http_s.ServerRef.WriteTimeout = 10 * time.Second
	http_s.ServerRef.IdleTimeout = 15 * time.Second
	http_s.ServerRef.ErrorLog = logRef.ERROR
	http_s.ServerRef.Handler = multiplexer
}

func (http_s *HttpServer) InitInterCommChannels(requestSendChl chan<- string, responseReceiveChl <-chan string) {
	http_s.SendRequestToUdp = requestSendChl
	http_s.RespnoseReceiveFromUdp = responseReceiveChl
}

func (http_s *HttpServer) Start(doneChannel chan<- bool) {

	http_s.Log_ref.Info("Http server Listening on ", http_s.ServerRef.Addr)
	if err := http_s.ServerRef.ListenAndServe(); err != http.ErrServerClosed {
		http_s.Log_ref.Error("Could not listen on specificed address : ", err)
	}

	doneChannel <- true
}
