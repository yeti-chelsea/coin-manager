package cmanager

import (
    "net/http"
)

const httpListenIp string = "0.0.0.0"

type HttpServer struct {
    // Port Number in which this HTTP server will be listening
    ListenPort int

    // Http server parameters
    ServerRef *http.Server

    // Pointer/reference to the logger interface
    Log_ref *Logger

    // A flag denoting running status
    Running bool
}

func (http *HttpServer) Init() {
}
