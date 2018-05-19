#!/usr/bin/env python3
'''
UDP client thread and socket
'''

import socket
import sys
import threading
from log_manager import Logger

HELLO_MSG = "Hello"
ACK_HELLO_MSG = "Ack-Hello"
KEEP_ALIVE = "Keep-Alive"
SEND_BASIC = "Send-Basic"

class UdpSocket(object):
    '''
    Simple wrapper around UDP socket
    '''
    def __init__(self, server_addr):
        self._sock_fd = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self._server_address = server_addr

    def udp_send(self, msg):
        '''
        Responsible for establishing a connection with the server
        '''
        return self._sock_fd.sendto(msg.encode(), self._server_address)

    def udp_receive(self, timeout=1):
        '''
        Responsible for receiving messages
        '''
        self._sock_fd.settimeout(timeout)
        return self._sock_fd.recvfrom(4096)


class UdpClientThread(threading.Thread):
    '''
    Simple UDP Client thread responsible for
    1. Connecting to c-manager server
    2. Sending basic info of the box like cpus, harware info
    3. Respond to keep alive when asked from server
    '''
    def __init__(self, server_addr, logger_ref):
        threading.Thread.__init__(self)
        self._logger_ref = logger_ref
        self._udp_client_interface = UdpSocket(server_addr)

    def run(self):
        '''
        Actual run method of the thread
        '''

        self._logger_ref.info("Starting thread...")
        # Initial Basic info initiated by client

        self._logger_ref.debug("Sending Basic info")
        bytes_sent = self._udp_client_interface.udp_send(HELLO_MSG)

        if bytes_sent < 1:
            self._logger_ref.critical("Failed to send packet, server might not be running")
            sys.exit(1)

        while True:

            try:
                data = self._udp_client_interface.udp_receive(6)
            except socket.timeout:
                self._logger_ref.warning("recv timed out, Sending hello message again")
                self._udp_client_interface.udp_send("hello")
                continue
            except socket.error:
                self._logger_ref.error("Socket exception error")
                sys.exit(1)
            else:
                actual_data = data[0].decode()

                if len(actual_data) == 0:
                    self._logger_ref.warning("Server has shutdown")
                    sys.exit(0)

                self._logger_ref.debug("Recevied message : ", actual_data)

                if actual_data == KEEP_ALIVE:
                    self._logger_ref.debug("Responding to keep alive.")
                    self._udp_client_interface.udp_send("i-m-alive")

                elif actual_data == SEND_BASIC:
                    self._logger_ref.debug("Resending basic info")
                    bytes_sent = self._udp_client_interface.udp_send("Hello")
                    if bytes_sent < 1:
                        self._logger_ref. \
                            critical("Failed to send packet, server might not be running")
                        sys.exit(1)

                elif actual_data == ACK_HELLO_MSG:
                    self._logger_ref.debug("Waiting for server to send the request")

                else:
                    self._logger_ref.warning("Unknown message received.")

'''
Testing
L_LOGGER = Logger("Ranjith", "stdout")
SERVER_ADD = ("localhost", 6767)
UDPCLIENT_MANAGER = UdpClientThread(SERVER_ADD, L_LOGGER)
UDPCLIENT_MANAGER.start()
'''
