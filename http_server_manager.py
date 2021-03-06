'''
Http server thread and socket
'''

import socket
import time
import threading
import json
from http.server import BaseHTTPRequestHandler, HTTPServer
import common_util

SUPPORTED_QUERY = ["mine-coin", "stop-mining", "mine-log", "current-mine-coin", "supported-query"]

def do_action(query, coin="", log_ref=None):
    '''
    Method to execute the query
    '''
    if query == "stop-mining":
        return common_util.stop_mining(log_ref)
    elif query == "mine-log":
        return common_util.get_mine_log(log_ref)
    elif query == "mine-coin":
        return common_util.start_mining(coin, log_ref)
    elif query == "current-mine-coin":
        return common_util.get_current_coin(log_ref)
    else:
        return "Query not supported"

class HTTPServerRequestHandler(BaseHTTPRequestHandler): # pylint:disable=too-few-public-methods
    '''
    Class responsible for handling HTTP request
    '''

    def __init__(self, logger_ref, *args):
        '''
        Initalize the logger
        '''
        self._logger_ref = logger_ref
        self._hostname = socket.gethostname() + " : "
        BaseHTTPRequestHandler.__init__(self, *args)

    def do_GET(self): # pylint: disable=invalid-name
        '''
        Do Get method impelemented
        '''

        self._logger_ref.debug("Received request from client : ", self.client_address)
        self._logger_ref.debug("Request Path : ", self.path)

        data = self.path
        response = self._hostname
        path = data.split('?')[0]
        query = data.split('?')[1]

        self._logger_ref.debug("Path : ", path)
        if path != "/rest/rproxy":
            self._logger_ref.warning("Unsupported path")
            response += "Unsupported url path"
            self._send_respose(response)
            return

        self._logger_ref.debug("Query : ", query)
        mine_query = ""
        if query.find('=') != -1:
            mine_query = query
            query = query.split('=')[0]

        supported_query = False
        for qry in SUPPORTED_QUERY:
            if query in qry:
                supported_query = True

        if not supported_query:
            self._logger_ref.warning("Unsupported Query : ", data)
            response += "Unsupported Query"
            self._send_respose(response)
            return

        if query == "supported-query":
            response += '/rest/rproxy?mine-coin=<coin-name>\n \
                    \r/rest/rproxy?stop-mining=\n \
                    \r/rest/rporxy?mine-log=\n \
                    \r/rest/rporxy?current-mine-coin=\n'
            self._send_respose(response)
            return

        if len(mine_query) > 1:
            response += do_action(query, coin=mine_query.split('=')[1], \
                    log_ref=self._logger_ref)
        else:
            response += do_action(query, log_ref=self._logger_ref)

        self._send_respose(response)

    def _send_respose(self, response):
        '''
        Method to send back the response
        '''
        self._logger_ref.debug("Sending : ", response)
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        json_string = json.dumps(response)
        self.wfile.write(json_string.encode())

class HttpServer(threading.Thread): # pylint:disable=too-few-public-methods
    '''
    A wrapper for Base Http Request handler
    '''
    def __init__(self, bind_address, log_ref):
        threading.Thread.__init__(self)
        def handler(*args):
            '''
            Handler function
            '''
            HTTPServerRequestHandler(log_ref, *args)

        self._logger_ref = log_ref
        self._http = HTTPServer(bind_address, handler)
        self._thread_start = True
        self._thread_running = False

    def run(self):
        '''
        Thread Run method
        '''
        self._thread_running = True
        try:
            self._http.serve_forever()
        except KeyboardInterrupt:
            pass
        except ValueError:
            pass

        self._thread_start = False

    def stop(self):
        '''
        On Calling stop we would close the server socket
        '''
        self._logger_ref.info("Shutting down HTTP server")
        self._thread_running = False

        self._http.server_close()
        
        while self._thread_start:
            time.sleep(1)

        self._logger_ref.info("HTTP server Down")
