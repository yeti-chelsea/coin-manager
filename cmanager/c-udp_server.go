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

const (
	CONSEQUTIVE_KEEP_ALIVE_TIMEOUT int = 3
	CONSEQUTIVE_SEND_FAILURES int = 3
	SLEEP_TIME_BEFORE_INTERACTING_INSEC int = 2
	TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC int = 4
	KEEP_ALIVE_SEND_TIMEOUT_INSEC int = 1
	SLEEP_TIME_AFTER_KEEP_ALIVE_TIMEOUT int = 2
)

type MinerStack struct {
	// Data receieved from specific client
	ClientDataChannel chan string

	// List of miner daemons supported
	MinerDaemons []string

	// List of supported coins
	Coins []string
}

type UdpServer struct {
	// Port Number in which this UDP server will be listening
	ListenPort int

	// Pointer to connection interface
	ConnRef *net.UDPConn

	// Pointer/reference to the logger interface
	Log_ref *Logger

	// Map of all the registered miners
	MapOfMiners map[string]*MinerStack

	// A flag denoting running status
	Running bool
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

func (udp *UdpServer) UdpClientGhoper(clientAddr <-chan *net.UDPAddr) {

	client_addr := <-clientAddr
	udp.Log_ref.Info("Launching a ghoper for client : ", client_addr.String())

	clientDataChl := udp.MapOfMiners[client_addr.String()].ClientDataChannel
	chl_data := <-clientDataChl
	udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",chl_data, client_addr))

	_,err := udp.ConnRef.WriteToUDP([]byte(ACK_HELLO_MSG), client_addr)
	if err != nil {
		udp.Log_ref.Error(err)
	}

	time.Sleep(time.Duration(SLEEP_TIME_BEFORE_INTERACTING_INSEC) * time.Second)
	var consecutiveKeepAliveTimeout int = 0
	var consecutiveFailures int = 0
	var breakout bool = false
	for udp.Running == true {

		udp.Log_ref.Debug("Sending keep alive message to : ", client_addr.String())
		_,err := udp.ConnRef.WriteToUDP([]byte(KEEP_ALIVE), client_addr)
		if err != nil {
			udp.Log_ref.Error(err)
			consecutiveFailures = consecutiveFailures + 1
			if consecutiveFailures == CONSEQUTIVE_SEND_FAILURES {
				udp.Log_ref.Warning("Consequtive send failures.. quitting !!!")
				break
			}
			time.Sleep(time.Duration(TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC) * time.Second)
			continue
		}
		consecutiveFailures = 0

		udp.Log_ref.Debug("Waiting for response from client : ", client_addr.String())
		select {
			case chl_data := <-clientDataChl :
				udp.Log_ref.Debug(fmt.Sprintf("Received message %s from client %s", chl_data, client_addr.String()))
				consecutiveKeepAliveTimeout = 0
				time.Sleep(time.Duration(TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC) * time.Second)
			case <-time.After(time.Duration(KEEP_ALIVE_SEND_TIMEOUT_INSEC) * time.Second):
				udp.Log_ref.Debug("Did not recieve response for keep-alive for 1 second")
				consecutiveKeepAliveTimeout = consecutiveKeepAliveTimeout + 1
				if consecutiveKeepAliveTimeout == CONSEQUTIVE_KEEP_ALIVE_TIMEOUT {
					udp.Log_ref.Warning("Keep alive timedout 3 consequtive times.. quitting !!!")
					breakout = true
					break
				}
				time.Sleep(time.Duration(SLEEP_TIME_AFTER_KEEP_ALIVE_TIMEOUT) * time.Second)
		}

		if breakout {
			break
		}
	}

	delete (udp.MapOfMiners, client_addr.String())
	udp.Log_ref.Warning("Ending this ghoper for client : ", client_addr.String())
}

func (udp *UdpServer) Start(doneChannel chan<- bool) {

	udp.Log_ref.Debug("Creating another channel for sending client Address")
	clAddr_channel := make(chan *net.UDPAddr, 1)

	buf := make([]byte, 1024)
	// Start the UDP server
	go func() {
		for udp.Running == true {
			udp.Log_ref.Debug("Waiting for messages in a while loop")
			bytes_read, client_addr, err := udp.ConnRef.ReadFromUDP(buf)
			if err != nil {
				udp.Log_ref.Error(err)
				continue
			}

			udp.Log_ref.Debug("Recieved data ")

			// Check whether this client is already registered with us 
			minerInfo, found := udp.MapOfMiners[client_addr.String()]

			if ! found {
				udp.Log_ref.Info("Registering new client : ", client_addr)
				udp.Log_ref.Debug("Creating new channel for this client")

				minerInfo = new(MinerStack)
				minerInfo.ClientDataChannel = make (chan string, 1)
				udp.MapOfMiners[client_addr.String()] = minerInfo
				go udp.UdpClientGhoper(clAddr_channel)
				clAddr_channel<- client_addr
			}

			minerInfo.ClientDataChannel<- string(buf[0:bytes_read])
		}
	}()

	doneChannel<- true
}
