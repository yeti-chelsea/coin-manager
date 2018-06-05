package cmanager

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"io/ioutil"
	"fmt"
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

	// Preferred Coin
	PreferredCoin string
}

func (http_s *HttpServer) ClientRegistered(registeredIp string) {
	http_s.Log_ref.Info(HTTP_LOGGER, "Got a notification from UDP ClientRegistered : ", registeredIp)

	if len(http_s.PreferredCoin) < 1 {
		http_s.Log_ref.Debug(HTTP_LOGGER, "Preferred Coin not set.. Not doing anything")
		return
	}

	res_from_miner := make(chan []byte, 1)

	coin_to_mine := "mine-coin=" + http_s.PreferredCoin

	go http_s.HttpClientRequest(registeredIp, coin_to_mine, res_from_miner)

	<-res_from_miner
}

func (http_s *HttpServer) ClientUnregistered(unregisteredIp string) {
	http_s.Log_ref.Info(HTTP_LOGGER, "Got a notification from UDP ClientUnregistered: ", unregisteredIp)
}

func (http_s *HttpServer) HttpClientRequest(minerHost string, request string, minerResponse chan<- []byte) {

	url := "http://" + strings.Split(minerHost, ":")[0] + ":6767" + "/rest/rproxy?" + request

    http_s.Log_ref.Debug(HTTP_LOGGER, "URL : ", url)

    res, err := http.Get(url)
    if err != nil {
        http_s.Log_ref.Warning(HTTP_LOGGER, err)
        minerResponse<- []byte{0}
        return
    }

    body, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        http_s.Log_ref.Warning(HTTP_LOGGER, err)
        minerResponse<- body
        return
    }

    http_s.Log_ref.Debug(HTTP_LOGGER, "Response received from client : ", minerHost, string(body))
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
	// "/rest/lserver?miner-host=all
	// "/rest/lserver?supported-query=all
	// "/rest/lserver?current-mine-coin=<all/miner-ip>
	http_s.Log_ref.Debug(HTTP_LOGGER, "Received request for serving locally : ", r.URL.RawQuery)

	supportedCurlRequest := []string {
		"miner-ip",
		"miner-coins",
		"miner-daemons",
		"stop-mining",
		"mine-log",
		"mine-coin",
		"miner-host",
		"current-mine-coin",
		"supported-query" }

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
	arg1 == supportedCurlRequest[2] ||
	arg1 == supportedCurlRequest[6] {
		http_s.Log_ref.Debug(HTTP_LOGGER, "Sending Request to UDP server")
		http_s.SendRequestToUdp<- []byte(r.URL.RawQuery)
		responseToClient = <-http_s.RespnoseReceiveFromUdp
	}

	if arg1 == supportedCurlRequest[3] ||
	arg1 == supportedCurlRequest[4] ||
	arg1 == "current-mine-coin" {
		http_s.Log_ref.Debug(HTTP_LOGGER, "Requesting for all miner ips from UDP")
		http_s.SendRequestToUdp<- []byte("miner-ip" + "=" + arg2)
		minerIpsbyteFormat := <-http_s.RespnoseReceiveFromUdp

		mIps := MIps{}
		json.Unmarshal(minerIpsbyteFormat, &mIps)

		if len(mIps.Ips) >= 1 {
			final_response := make([]string, len(mIps.Ips))
			for index, reg_ip := range mIps.Ips {
				res_from_miner := make(chan []byte, 1)
				http_s.Log_ref.Debug(HTTP_LOGGER, fmt.Sprintf("Requesting client : %v for %v", reg_ip, arg1))
				go http_s.HttpClientRequest(reg_ip, arg1, res_from_miner)

				response := <-res_from_miner
				http_s.Log_ref.Debug(HTTP_LOGGER, fmt.Sprintf("Received response from client : %v : %v", reg_ip, string(response)))
				final_response[index] = string(response)
			}

			responseToClientStr := strings.Join(final_response, "\n")
			responseToClient = []byte(responseToClientStr)
		}
	}

	if arg1 == supportedCurlRequest[5] {
		// mineIp := strings.Split(arg2, "?")[0]
		coin := strings.Split(arg2, "?")[1]

		http_s.Log_ref.Info(HTTP_LOGGER, "Setting preferred coin : ", coin)
		http_s.PreferredCoin = coin

		http_s.Log_ref.Debug(HTTP_LOGGER, "Requesting for all miner ips from UDP")
		http_s.SendRequestToUdp<- []byte("miner-ip" + "=" + "all")
		minerIpsbyteFormat := <-http_s.RespnoseReceiveFromUdp

		mIps := MIps{}
		json.Unmarshal(minerIpsbyteFormat, &mIps)

		if len(mIps.Ips) < 1 {
			http_s.Log_ref.Debug(HTTP_LOGGER, "No miner have registered yet, just setting preferred coin")
			responseToClient = []byte("Preferred Coin set")
		}else {
			final_response := make([]string, len(mIps.Ips))
			for index, reg_ip := range mIps.Ips {
				res_from_miner := make(chan []byte, 1)
				http_s.Log_ref.Debug(HTTP_LOGGER, fmt.Sprintf("Requesting client : %v for %v", reg_ip, arg1))
				go http_s.HttpClientRequest(reg_ip, arg1 + "=" + coin, res_from_miner)

				response := <-res_from_miner
				http_s.Log_ref.Debug(HTTP_LOGGER, fmt.Sprintf("Received response from client : %v : %v", reg_ip, string(response)))
				final_response[index] = string(response)
			}

			responseToClientStr := strings.Join(final_response, "\n")
			responseToClient = []byte(responseToClientStr)
		}
	}

	if arg1 == "supported-query" {
		var supportedQuery string
		supportedQuery += "/rest/lserver?miner-ip=<all/miner-ip>\n" +
		"\r/rest/lserver?miner-coins=<all/miner-ip>\n" +
		"\r/rest/lserver?miner-daemons=<all/miner-ip>\n" +
		"\r/rest/lserver?mine-coin=<all/miner-ip>?<coin>\n" +
		"\r/rest/lserver?stop-mining=<all/miner-ip>\n" +
		"\r/rest/lserver?mine-log=<all/miner-ip>\n" +
		"\r/rest/lserver?miner-host=all\n" +
		"\r/rest/lserver?current-mine-coin=<all/miner-ip>\n"
		responseToClient = []byte(supportedQuery)
	}

	http_s.Log_ref.Debug(HTTP_LOGGER, "Response sent  : ", string(responseToClient))
	w.Write(responseToClient)
}

func (http_s *HttpServer) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Supported URL's
	// "/rest/rproxy?mine-coin=<coin-name>"
	// "/rest/rproxy?stop-mining"
	// "/rest/rproxy?mine-log"

	http_s.Log_ref.Debug(HTTP_LOGGER, "Received request for proxying")
	http_s.Log_ref.Debug(HTTP_LOGGER, r.URL.Path)
}

func (http_s *HttpServer) AddHandlers(pattern string, handler myHandlers) {
	http_s.Controller[pattern] = handler
}

func (http_s *HttpServer) Init(listenIp string, listenPort int, logRef *Logger) {

	http_s.Log_ref = logRef
	http_s.Log_ref.SetLogLevel(HTTP_LOGGER, INFO_LEVEL)
	multiplexer := http.NewServeMux()

	http_s.Controller = make(map[string]myHandlers)

	http_s.Log_ref.Debug(HTTP_LOGGER, "Initializing HTTP server")
	http_s.AddHandlers("/rest/lserver", http_s.LocalRequestHandler)
	http_s.AddHandlers("/rest/rproxy", http_s.ProxyRequestHandler)

	for pattern, request_handler := range http_s.Controller {
		http_s.Log_ref.Debug(HTTP_LOGGER, "Registering handler with multiplexer : ", pattern)
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

	http_s.Log_ref.Info(HTTP_LOGGER, "Http server Listening on ", http_s.ServerRef.Addr)
	if err := http_s.ServerRef.ListenAndServe(); err != http.ErrServerClosed {
		http_s.Log_ref.Error(HTTP_LOGGER, "Could not listen on specificed address : ", err)
	}

	doneChannel <- true
}
