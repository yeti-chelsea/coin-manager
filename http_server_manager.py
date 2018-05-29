#!/usr/bin/env python3

'''
Http server thread and socket
'''

import socket
import threading
import time
from urllib.parse import urlparse

SUPPORTED_QUERY = ["mine-coin", "stop-mining", "mine-log", "current-mine-coin"]

def do_action():
    '''

    '''

class HttpServerSocket(object):
    '''
    Simple wrapper around HTTP/TCP socket
    '''
    def __init__(self, address='0.0.0.0:6767'):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._bind_address = address

        self._socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
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

            self._logger_ref.info("Received connection from client : ", client_addr)
            data = client_socket.recv(1024)
            response = ""
            url_data = urlparse(data.decode())
            path = url_data.path.split(' ')[1]
            query = url_data.query.split(' ')[0]

            if path != "/rest/rproxy":
                self._logger_ref.warning("Unsupported path")
                response = "Unsupported url path"
                client_socket.sendall(response.encode())
                client_socket.close()
                continue

            mine_query = ""
            if query.find('=') != -1:
                mine_query = query
                query = query.split('=')[0]

            supported_query = False
            for qry in SUPPORTED_QUERY:
                if query in qry:
                    supported_query = True

            if supported_query == False:
                response = "Unsupported Query"
                client_socket.sendall(response.encode())
                client_socket.close()
                continue

            response = "Success"
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
