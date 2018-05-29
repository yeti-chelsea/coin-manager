package cmanager

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"io/ioutil"
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
	SendRequestToUdp chan<- []byte

	// Channel for receiving Response from UDP server
	RespnoseReceiveFromUdp <-chan []byte
}

func (http_s *HttpServer) HttpClientRequest(minerHost string, request string, minerResponse chan<- []byte) {

	url := "http://" + minerHost
	payload := strings.NewReader(request)

	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	minerResponse<- body

}

func (http_s *HttpServer) LocalRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Supported URL's
	// "/rest/lserver?miner-ip=<all/miner-ip>
	// "/rest/lserver?miner-coins=<all/miner-ip>
	// "/rest/lserver?miner-daemons=<all/miner-ip>
	// "/rest/lserver?mine-coin=<all/miner-ip>?<coin>"
	// "/rest/lserver?stop-mining=<all/miner-ip>"
	// "/rest/lserver?mine-log=<all/miner-ip>
	http_s.Log_ref.Debug("Received request for serving locally : ", r.URL.RawQuery)

	supportedCurlRequest := []string {
		"miner-ip",
		"miner-coins",
		"miner-daemons",
		"stop-mining",
		"mine-log",
		"mine-coin" }

	responseToClient := []byte("Unsupported-Query")

	var requestFound = false
	for _,supported_req := range supportedCurlRequest {
		if strings.Contains(r.URL.RawQuery, supported_req) {
			requestFound = true
			break
		}
	}

	if ! requestFound {
		w.Write(responseToClient)
		return
	}

	arg1 := strings.Split(r.URL.RawQuery, "=")[0]
	arg2 := strings.Split(r.URL.RawQuery, "=")[1]

	if arg1 == supportedCurlRequest[0] ||
	arg1 == supportedCurlRequest[1] ||
	arg1 == supportedCurlRequest[2] {
		http_s.Log_ref.Debug("Sending Request to UDP server")
		http_s.SendRequestToUdp<- []byte(r.URL.RawQuery)
		responseToClient = <-http_s.RespnoseReceiveFromUdp
		http_s.Log_ref.Debug("Received response from UDP server : ", responseToClient)
	}

	if arg1 == supportedCurlRequest[3] ||
	arg1 == supportedCurlRequest[4] {
		http_s.Log_ref.Debug("Requesting for all miner ips")
		http_s.SendRequestToUdp<- []byte("miner-ip" + "=" + arg2)
		minerIpsbyteFormat := <-http_s.RespnoseReceiveFromUdp

		mIps := MIps{}
		json.Unmarshal(minerIpsbyteFormat, mIps)

		for _, reg_ips := range mIps.Ips {
			res_from_miner := make(chan []byte, 1)
			go http_s.HttpClientRequest(reg_ips, arg1, res_from_miner)

			// TODO : Response must be multiplexed from all the gophers
			responseToClient = <-res_from_miner
		}

		http_s.Log_ref.Debug("Response sent  : ", responseToClient)
	}

	if arg1 == supportedCurlRequest[5] {
		mineIp := strings.Split(arg2, "?")[0]
		coin := strings.Split(arg2, "?")[1]

		http_s.Log_ref.Debug("Requesting for all miner ips")
		http_s.SendRequestToUdp<- []byte("miner-ip" + "=" + mineIp)
		minerIpsbyteFormat := <-http_s.RespnoseReceiveFromUdp

		mIps := MIps{}
		json.Unmarshal(minerIpsbyteFormat, mIps)

		for _, reg_ips := range mIps.Ips {
			res_from_miner := make(chan []byte, 1)
			go http_s.HttpClientRequest(reg_ips, arg1 + "=" + coin, res_from_miner)

			// TODO : Response must be multiplexed from all the gophers
			responseToClient = <-res_from_miner
		}
		http_s.Log_ref.Debug("Response sent  : ", responseToClient)
	}

	w.Write(responseToClient)
}

func (http_s *HttpServer) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Supported URL's
	// "/rest/rproxy?mine-coin=<coin-name>"
	// "/rest/rproxy?stop-mining"
	// "/rest/rproxy?mine-log"

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

func (http_s *HttpServer) InitInterCommChannels(requestSendChl chan<- []byte, responseReceiveChl <-chan []byte) {
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
