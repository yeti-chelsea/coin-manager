#!/usr/bin/env python3

'''
Http server thread and socket
'''
# https://cpiekarski.com/2011/05/09/super-easy-python-json-client-server/

import socket
import threading
import time
from urllib.parse import urlparse

class HttpServerSocket(object):
    '''
    Simple wrapper around HTTP/TCP socket
    '''
    def __init__(self, address='0.0.0.0:6767'):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._bind_address = address
        self._socket.bind(self._bind_address)
        self._socket.settimeout(3)
        self._socket.listen(2)

    def accept_connection(self):
        '''
        Wrapper for accepting connection
        '''
        return self._socket.accept()

    def close(self):
        '''
        Wrapper to close the server socket
        '''
        self._socket.close()

class HttpServerThread(threading.Thread):
    '''
    Thread responsible for receiving rest call and responding
    a json object

    Following rest call are supported
    /rest/rproxy?mine-coin=<coin-name>
    /rest/rproxy?stop-mining=
    /rest/rporxy?mine-log=
    /rest/rporxy?current-mine-coin=
    '''
    def __init__(self, bind_addr, logger_ref):
        threading.Thread.__init__(self)
        self._logger_ref = logger_ref
        self.http_server_socket = HttpServerSocket(bind_addr)
        self._logger_ref.debug("Initalizing HTTP socket")
        self._thread_start = True
        self._thread_runnung = False

    def run(self):
        self._logger_ref.debug("Starting HTTP server")
        self._thread_runnung = True
        while self._thread_runnung:
            try:
                client_socket, client_addr = \
                self.http_server_socket.accept_connection()
            except socket.timeout:
                continue

            print("Received connection from client : ", client_addr)
            data = client_socket.recv(1024)
            #print(data.decode())
            url_data = urlparse(data.decode())
            print("Compelte : ", url_data)
            print("Path", url_data.path[1])
            print("Query", url_data.query)
            print("Param", url_data.params)
            response = "Accepted"
            client_socket.sendall(response.encode())
            client_socket.close()

        self._thread_start = False

    def stop(self):
        '''
        On Calling stop we would close the server socket
        '''
        self._logger_ref.info("Shutting down HTTP server")
        self._thread_runnung = False

        while self._thread_start:
            time.sleep(1)

        self._logger_ref.debug("Closing server socket")
        self.http_server_socket.close()
        self._logger_ref.info("HTTP server Down")
