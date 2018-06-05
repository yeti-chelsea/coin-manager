'''
Http server thread and socket
'''

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
        BaseHTTPRequestHandler.__init__(self, *args)

    def do_GET(self): # pylint: disable=invalid-name
        '''
        Do Get method impelemented
        '''

        self._logger_ref.debug("Received request from client : ", self.client_address)
        self._logger_ref.debug("Request Path : ", self.path)

        data = self.path
        response = ""
        path = data.split('?')[0]
        query = data.split('?')[1]

        self._logger_ref.debug("Path : ", path)
        if path != "/rest/rproxy":
            self._logger_ref.warning("Unsupported path")
            response = "Unsupported url path"
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
            response = "Unsupported Query"
            self._send_respose(response)
            return

        if query == "supported-query":
            response = '/rest/rproxy?mine-coin=<coin-name>\n \
                    \r/rest/rproxy?stop-mining=\n \
                    \r/rest/rporxy?mine-log=\n \
                    \r/rest/rporxy?current-mine-coin\n'
            self._send_respose(response)
            return

        if len(mine_query) > 1:
            response = do_action(query, coin=mine_query.split('=')[1], \
                    log_ref=self._logger_ref)
        else:
            response = do_action(query, log_ref=self._logger_ref)

        self._send_respose(response)

    def _send_respose(self, response):
        '''
        Method to send back the response
        '''
        self.wfile.write(bytes(response, "utf8"))

class HttpServer(object): # pylint:disable=too-few-public-methods
    '''
    A wrapper for Base Http Request handler
    '''
    def __init__(self, bind_address, log_ref):
        def handler(*args):
            '''
            Handler function
            '''
            HTTPServerRequestHandler(log_ref, *args)

        httpd = HTTPServer(bind_address, handler)
        httpd.serve_forever()
