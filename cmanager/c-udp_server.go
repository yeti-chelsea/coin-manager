package cmanager

import (
	"fmt"
	"net"
	"time"
)

const listenIp string = "0.0.0.0"

const (
	HELLO_MSG string = "Hello"
	ACK_HELLO_MSG string = "Ack-Hello"
	KEEP_ALIVE string = "Keep-Alive"
	SEND_BASIC string = "Send-Basic"
)

type UdpServer struct {
	ListenPort int

	ConnRef *net.UDPConn
	Log_ref *Logger
	MapOfMiners map[string]chan string
}

func (udp *UdpServer) Init() error {
	// Initalize the UDP server
	addr := net.UDPAddr{
        Port: udp.ListenPort,
        IP: net.ParseIP(listenIp),
    }

	var err error
	udp.Log_ref.Info(fmt.Sprintf("Listering on ip : %v and port : %v", listenIp, udp.ListenPort))
	udp.ConnRef, err = net.ListenUDP("udp", &addr)

	if err != nil {
		return err
	}

	return nil
}

func (udp *UdpServer) ManageClientsMessages(clientAddr <-chan *net.UDPAddr) {

	for {
		client_addr := <-clientAddr
		chl_data := <-udp.MapOfMiners[client_addr.String()]
		udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",chl_data, client_addr))

		_,err := udp.ConnRef.WriteToUDP([]byte(ACK_HELLO_MSG), client_addr)

		if err != nil {
			udp.Log_ref.Error(err)
		}
	}
}

func (udp *UdpServer) UdpClientGhoper(clientAddr <-chan *net.UDPAddr) {

	client_addr := <-clientAddr
	udp.Log_ref.Info("Launching a ghoper for client : ", client_addr.String())
	chl_data := <-udp.MapOfMiners[client_addr.String()]
	udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",chl_data, client_addr))

	_,err := udp.ConnRef.WriteToUDP([]byte(ACK_HELLO_MSG), client_addr)
	if err != nil {
		udp.Log_ref.Error(err)
	}

	time.Sleep(3 * time.Second)
	var consecutiveKeepAliveTimeout int
	var consecutiveFailures int
	var breakout bool = false
	for {

		udp.Log_ref.Debug("Sending keep alive message to : ", client_addr.String())
		_,err := udp.ConnRef.WriteToUDP([]byte(KEEP_ALIVE), client_addr)
		if err != nil {
			udp.Log_ref.Error(err)
			consecutiveFailures = consecutiveFailures + 1
			if consecutiveFailures == 3 {
				udp.Log_ref.Warning("Consequtive send failures.. quitting !!!")
				break
			}
		}
		consecutiveFailures = 0

		udp.Log_ref.Debug("Waiting for response from client : ", client_addr.String())
		select {
			case chl_data := <-udp.MapOfMiners[client_addr.String()] :
				udp.Log_ref.Debug(fmt.Sprintf("Received message %s from client %s", chl_data, client_addr.String()))
				consecutiveKeepAliveTimeout = 0
				time.Sleep(4 * time.Second)
			case <-time.After(1 * time.Second):
				udp.Log_ref.Debug("Did not recieve response for keep-alive for 1 second")
				consecutiveKeepAliveTimeout = consecutiveKeepAliveTimeout + 1
				if consecutiveKeepAliveTimeout == 3 {
					udp.Log_ref.Warning("Keep alive timedout 3 consequtive times.. quitting !!!")
					breakout = true
					break
				}
				time.Sleep(2 * time.Second)
				/*
			default:
				udp.Log_ref.Warning("Unknown")
				time.Sleep(4 * time.Second)
				*/
		}

		if breakout {
			break
		}
	}

	delete (udp.MapOfMiners, client_addr.String())
	udp.Log_ref.Warning("Ending this ghoper for client : ", client_addr.String())
}

func (udp *UdpServer) Start(doneChannel chan<- bool) {

	udp.Log_ref.Debug("Creating another channel cl_channel")
	cl_channel := make(chan *net.UDPAddr, 1)

	buf := make([]byte, 1024)
	// Start the UDP server
	go func() {
		for {
			udp.Log_ref.Debug("Waiting for messages in a while loop")
			bytes_read, client_addr, err := udp.ConnRef.ReadFromUDP(buf)
			if err != nil {
				udp.Log_ref.Error(err)
				continue
			}

			// convert the data into string
			data := string(buf[0:bytes_read])
			udp.Log_ref.Debug("Recieved data ")

			// Check whether this client is already registered with us 
			client_ch, found := udp.MapOfMiners[client_addr.String()]

			if ! found {
				udp.Log_ref.Info("Registering new client : ", client_addr)
				udp.Log_ref.Debug("Creating new channel for this client")
				udp.MapOfMiners[client_addr.String()] = make (chan string, 1)
				client_ch = udp.MapOfMiners[client_addr.String()]
				go udp.UdpClientGhoper(cl_channel)
				cl_channel<- client_addr
			}

			client_ch<- data
		}
	}()

	doneChannel<- true
}
