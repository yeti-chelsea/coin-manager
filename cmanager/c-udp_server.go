package cmanager

import (
	"fmt"
	"net"
	"time"
    "encoding/json"
)

const (
	HELLO_MSG string = "Hello"
	ACK_HELLO_MSG string = "Ack-Hello"
	KEEP_ALIVE string = "Keep-Alive"
	SEND_BASIC string = "Send-Basic"
	MINER_DAEMONS string = "Miner-Daemons"
	MINER_COINS string = "Miner-Coins"
)

const (
	CONSEQUTIVE_KEEP_ALIVE_TIMEOUT int = 3
	CONSEQUTIVE_SEND_FAILURES int = 3
	SLEEP_TIME_BEFORE_INTERACTING_INSEC int = 2
	TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC int = 4
	SEND_TIMEOUT_INSEC int = 1
	SLEEP_TIME_AFTER_KEEP_ALIVE_TIMEOUT int = 2
)

type MDaemons struct {
    Daemons []string `json:"miner-daemons"`
}

type MCoins struct {
    Coins []string `json:"miner-coins"`
}

type MinerStack struct {
	// Channel for receiving data from specific client
	ClientDataChannel chan []byte

	// List of miner daemons supported
	MinerDaemons *MDaemons

	// List of supported coins
	MinerCoins *MCoins
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

func (udp *UdpServer) Init(listenIp string, listenPort int, logRef *Logger) error {
	// Initalize the UDP server

    udp.ListenPort = listenPort
    udp.Log_ref = logRef
    udp.MapOfMiners = make(map[string]*MinerStack)

	addr := net.UDPAddr{
        Port: udp.ListenPort,
        IP: net.ParseIP(listenIp),
    }

	var err error
	udp.Log_ref.Info(fmt.Sprintf("UDP server Listering on %v:%v", listenIp, udp.ListenPort))
	udp.ConnRef, err = net.ListenUDP("udp", &addr)

	if err != nil {
		return err
	}

    udp.Running = true
	return nil
}

func (udp *UdpServer) UdpClientGhoper(clientAddr <-chan *net.UDPAddr) {

	client_addr := <-clientAddr
	udp.Log_ref.Info("Launching a ghoper for client : ", client_addr.String())

    minerStack := udp.MapOfMiners[client_addr.String()]
	clientDataChl := minerStack.ClientDataChannel

	chl_data_bytes := <-clientDataChl
	udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",string(chl_data_bytes), client_addr))

	_,err := udp.ConnRef.WriteToUDP([]byte(ACK_HELLO_MSG), client_addr)
	if err != nil {
		udp.Log_ref.Error(err)
	}

	time.Sleep(time.Duration(SLEEP_TIME_BEFORE_INTERACTING_INSEC) * time.Second)

	udp.Log_ref.Debug("Asking for Miner Daemons")
	_,err = udp.ConnRef.WriteToUDP([]byte(MINER_DAEMONS), client_addr)
	if err != nil {
		udp.Log_ref.Error(err)
	}

	select {
		case chl_data_bytes = <-clientDataChl:
			udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",string(chl_data_bytes), client_addr))
		case <-time.After(time.Duration(SEND_TIMEOUT_INSEC) * time.Second):
			udp.Log_ref.Debug(fmt.Sprintf("UDP send timedout after waiting for %v seconds", SEND_TIMEOUT_INSEC))
			delete (udp.MapOfMiners, client_addr.String())
			udp.Log_ref.Warning("Ending this ghoper for client : ", client_addr.String())
			return
	}

    json.Unmarshal(chl_data_bytes, minerStack.MinerDaemons)
    udp.Log_ref.Debug(minerStack.MinerDaemons)

	time.Sleep(time.Duration(SLEEP_TIME_BEFORE_INTERACTING_INSEC) * time.Second)

	udp.Log_ref.Debug("Asking for Miner Coins")
	_,err = udp.ConnRef.WriteToUDP([]byte(MINER_COINS), client_addr)
	if err != nil {
		udp.Log_ref.Error(err)
	}

	select {
		case chl_data_bytes = <-clientDataChl:
			udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",string(chl_data_bytes), client_addr))
		case <-time.After(time.Duration(SEND_TIMEOUT_INSEC) * time.Second):
			udp.Log_ref.Debug(fmt.Sprintf("UDP send timedout after waiting for %v seconds", SEND_TIMEOUT_INSEC))
			delete (udp.MapOfMiners, client_addr.String())
			udp.Log_ref.Warning("Ending this ghoper for client : ", client_addr.String())
			return
	}


    json.Unmarshal(chl_data_bytes, minerStack.MinerCoins)
    udp.Log_ref.Debug(minerStack.MinerCoins)

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
			case chl_data_bytes := <-clientDataChl :
				udp.Log_ref.Debug(fmt.Sprintf("Received message %s from client %s", string(chl_data_bytes), client_addr.String()))
				consecutiveKeepAliveTimeout = 0
				time.Sleep(time.Duration(TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC) * time.Second)
			case <-time.After(time.Duration(SEND_TIMEOUT_INSEC) * time.Second):
				udp.Log_ref.Info("Did not recieve response for keep-alive for")
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
	udp.Log_ref.Info("Ending this ghoper for client : ", client_addr.String())
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
				minerInfo.ClientDataChannel = make (chan []byte, 1)
                minerInfo.MinerDaemons = new(MDaemons)
                minerInfo.MinerCoins = new(MCoins)
				udp.MapOfMiners[client_addr.String()] = minerInfo
				go udp.UdpClientGhoper(clAddr_channel)
				clAddr_channel<- client_addr
			}

			minerInfo.ClientDataChannel<- buf[0:bytes_read]
		}
	}()

	doneChannel<- true
}
