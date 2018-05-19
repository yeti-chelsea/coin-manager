package cmanager

import (
	"fmt"
	"net"
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
	MapOfMiners map[net.IP]chan string
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
		chl_data := <-udp.MapOfMiners[client_addr.IP]
		udp.Log_ref.Debug(fmt.Sprintf("Received message %s from %s",chl_data, client_addr))

		_,err := udp.ConnRef.WriteToUDP([]byte(ACK_HELLO_MSG), client_addr)

		if err != nil {
			udp.Log_ref.Error(err)
		}
	}

}

func (udp *UdpServer) Start(doneChannel chan<- bool) {
	//defer udp.ConnRef.Close()

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

			udp.Log_ref.Debug("Lenth of map : ", len(udp.MapOfMiners))
			// Check whether this client is already registered with us 
			client_ch, found := udp.MapOfMiners[client_addr.IP]

			if ! found {
				udp.Log_ref.Info("Registering new client : ", client_addr)
				udp.Log_ref.Debug("Creating new channel for this client")
				udp.MapOfMiners[client_addr.IP] = make (chan string, 1)
				client_ch = udp.MapOfMiners[client_addr.IP]
			}

			udp.Log_ref.Debug(udp.MapOfMiners)

			client_ch<- data
			cl_channel<- client_addr
		}
	}()

	go udp.ManageClientsMessages(cl_channel)
	doneChannel<- true
}
