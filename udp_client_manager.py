'''
UDP client thread and socket
'''

import socket
import sys
import threading
import time
import common_util

HELLO_MSG = "Hello"
ACK_HELLO_MSG = "Ack-Hello"
KEEP_ALIVE = "Keep-Alive"
SEND_BASIC = "Send-Basic"
MINER_DAEMONS = "Miner-Daemons"
MINER_COINS = "Miner-Coins"
HOST_NAME = "Host-Name"

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

    def close_socket(self):
        '''
        Wrapper for closing the socket
        '''
        self._sock_fd.close()

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
        self._thread_start = True
        self._thread_runnung = True

    def run(self):
        '''
        Actual run method of the thread
        '''

        self._logger_ref.info("Starting thread...")
        # Initial Basic info initiated by client

        self._logger_ref.debug("Sending Hello message to the server")
        bytes_sent = self._udp_client_interface.udp_send(HELLO_MSG)

        if bytes_sent < 1:
            self._logger_ref.critical("Failed to send packet, server might not be running")
            sys.exit(1)

        while self._thread_runnung:

            try:
                data = self._udp_client_interface.udp_receive(6)
            except socket.timeout:
                self._logger_ref.warning("recv timed out, Sending hello message again")
                self._udp_client_interface.udp_send("hello")
                continue
            except socket.error:
                self._logger_ref.error("Socket exception error")
                sys.exit(1)

            # Data came on wire will be in byte format
            actual_data = data[0].decode()

            if len(actual_data) == 0:
                self._logger_ref.warning("Server has shutdown")
                sys.exit(0)

            self._logger_ref.debug("Receivied message : ", actual_data)

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

            elif actual_data == MINER_DAEMONS:
                self._logger_ref.debug("Sending Miner Daemons Info")
                bytes_sent = \
                self._udp_client_interface.udp_send(common_util.get_miner_daemons_json())

            elif actual_data == MINER_COINS:
                self._logger_ref.debug("Sending Coins supported by miners")
                bytes_sent = \
                self._udp_client_interface.udp_send(common_util.get_miner_coins_json())

            elif actual_data == ACK_HELLO_MSG:
                self._logger_ref.debug("Waiting for server to send the request")

            elif actual_data == HOST_NAME:
                self._logger_ref.debug("Sending host-name")
                self._udp_client_interface.udp_send(socket.gethostname().encode())

            else:
                self._logger_ref.warning("Unknown message received.")

        self._thread_start = False

    def stop(self):
        '''
        Stop the thread
        '''
        self._logger_ref.info("Shutting down UDP server")
        self._thread_runnung = False

        while self._thread_start:
            time.sleep(1)

        self._logger_ref.debug("Closing server socket")
        self._udp_client_interface.close_socket()
        self._logger_ref.info("UDP server down")
