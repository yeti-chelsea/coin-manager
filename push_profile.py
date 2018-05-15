#!/usr/bin/env python3

'''
A simple script which automates profile pushes
'''

import json
import os
import getopt
import sys
import time
from copy import deepcopy
from http.server import BaseHTTPRequestHandler, HTTPServer
import paramiko
from log_manager import Logger

L_LOGGER_FILE = None

class SshConnection(object):
    '''
    A simple class responsible for making a remote
    ssh connection for the given host-name and user
    '''
    def __init__(self, hostname, username, password):
        self.__hostname = hostname
        self.__username = username
        self.__password = password
        try:
            self.__host_keys = \
                    paramiko.util.load_host_keys(os.path.expanduser('~/.ssh/known_hosts'))
        except IOError:
            L_LOGGER_FILE.critical("Failed opening the file : ~/.ssh/known_hosts")

        self.__ssh_client = paramiko.SSHClient()

    def connect(self):
        '''
        Method which makes the actuall connection
        to the ssh server
        '''
        rsa_pub_key = self.__host_keys[self.__hostname]['ssh-rsa']

        self.__ssh_client.get_host_keys().add(self.__hostname, 'ssh-rsa', rsa_pub_key)
        self.__ssh_client.connect(self.__hostname, \
                username=self.__username, \
                password=self.__password)

    def execute_cmd(self, cmd):
        '''
        Method responsible for executing remote command
        '''
        self.__ssh_client.exec_command(cmd)

    def execute_cmd_and_exit(self, cmd):
        '''
        Method responsible for executing remote command
        and exit
        '''
        self.__ssh_client.exec_command(cmd)
        self.__ssh_client.close()

    def close_connection(self):
        '''
        Method responsible for closing the connection
        '''
        self.__ssh_client.close()

def make_full_command(json_dictinary):
    '''
    This method is responsible for making the complete
    command which will be executed in the remote machine
    '''

    local_copy_json_dict = deepcopy(json_dictinary)
    screen_name = local_copy_json_dict['screen-name']
    loc = local_copy_json_dict['location']
    del local_copy_json_dict['screen-name']
    del local_copy_json_dict['location']

    screen_cmd = "screen"
    bash_option = "sh -c"

    tcl_path = ""

    if loc == "blr":
        tcl_path = "/m/test_main/fwtest/bin/tcl"
    elif loc == "wtl":
        tcl_path = "/m/test_main/fwtest/bin/tcl"
    else:
        tcl_path = "/m/test_main/fwtest/bin/tcl"

    remaining_attr = ""
    for key, val in local_copy_json_dict.items():
        remaining_attr += "-" + str(key) + " " + str(val) + " "

    full_cmd = screen_cmd +  " -dmSL" + " " + screen_name + " " + \
            bash_option + " '" + tcl_path + " profilemgr " + remaining_attr + "'"

    return full_cmd

# HTTPRequestHandler class
class HTTPServerRequestHandler(BaseHTTPRequestHandler): # pylint:disable=too-few-public-methods
    '''
    Http Server for handling imcoming request
    '''

    USERNAME = None
    PASSWD = None
    BLR_TEST_MACHINE = "blrlab-test-2"
    WTL_TEST_MACHINE = "wtllab-test-10"

    PREVIOUS_TIME_STAMP_BLR = 0
    PREVIOUS_TIME_STAMP_WTL = 0
    ACK_MESSAGE = "Acked\n"
    # GET
    def do_GET(self): # pylint: disable=invalid-name
        '''
        Do Get method impelemented
        '''
        # Get the current time stamp
        current_time_stamp = int(time.time())

        kill_screen = 0
        # Get the length of the message
        msg_len = int(self.headers['Content-Length'])

        #Get the request data
        get_data = self.rfile.read(msg_len).decode()
        L_LOGGER_FILE.debug(get_data)

        # Get the request data in json object
        loaded_json = json.loads(get_data)

        # Get all the attributes of json in a dictinary
        profile_dict = {}
        for attr in loaded_json:
            profile_dict[attr] = loaded_json[attr]

        cmd = make_full_command(profile_dict)
        L_LOGGER_FILE.info("Complete command to execute : ", cmd)

        h_name = ""
        loc = profile_dict.get("location")

        if loc == "blr":
            h_name = HTTPServerRequestHandler.BLR_TEST_MACHINE

            L_LOGGER_FILE.debug("Current time, Previous blr time, diffrence ", \
                    current_time_stamp, HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_BLR, \
                    (current_time_stamp - HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_BLR))
            if (current_time_stamp - \
                    HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_BLR) > 1800:
                kill_screen = 1

            HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_BLR = current_time_stamp
        elif loc == "wtl":
            h_name = HTTPServerRequestHandler.WTL_TEST_MACHINE
            L_LOGGER_FILE.debug("Current time, Previous wtl time, diffrence ", \
                    current_time_stamp, HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_WTL, \
                    (current_time_stamp - HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_WTL))

            if (current_time_stamp - HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_WTL) > 1800:
                kill_screen = 1
            HTTPServerRequestHandler.PREVIOUS_TIME_STAMP_WTL = current_time_stamp
        else:
            L_LOGGER_FILE.error("Unknow location received")
            return

        L_LOGGER_FILE.debug("Creating SshConnection object")
        ssh_c = SshConnection(h_name, \
                HTTPServerRequestHandler.USERNAME, \
                HTTPServerRequestHandler.PASSWD)

        L_LOGGER_FILE.info("Connecting to remote server : ", h_name)
        ssh_c.connect()

        if kill_screen == 1:
            L_LOGGER_FILE.debug("Sending kilall screen command")
            ssh_c.execute_cmd("killall screen")

        L_LOGGER_FILE.debug("Executing remote command and exiting")
        ssh_c.execute_cmd_and_exit(cmd)

        L_LOGGER_FILE.debug("Sending Ack back to the client : ", \
                self.client_address)
        # Write content as utf-8 data
        self.wfile.write(bytes(HTTPServerRequestHandler.ACK_MESSAGE, "utf8"))
        return

def run(port_number):
    '''
    Main Run method
    '''
    L_LOGGER_FILE.info('starting server...')

    # Server settings
    server_address = ('0.0.0.0', port_number)
    httpd = HTTPServer(server_address, HTTPServerRequestHandler)
    L_LOGGER_FILE.debug('running server...')
    httpd.serve_forever()

def usage():
    '''
    Print usage
    '''
    print("./push_profile.py \n\
            -h --help : display help message \n\
            -u --username : Username to ssh \n\
            -p --password : password of the user \n\
            -n --port : port number to bind \n\
            -l --logfile : log file to log\n")
    sys.exit()

try:
    OPTS, ARGS = getopt.getopt(sys.argv[1:], "hu:p:l:n:", ["help", "username=", \
        "password=", "logfile=", "port="])
except getopt.GetoptError as err:
    print(err)
    usage()
    sys.exit(2)

U_NAME = None
P_WD = None
LOG_FILE = None
PORT_NUMBER = 8081

for opt, value in OPTS:
    if opt in ("-u", "--username"):
        U_NAME = value
    elif opt in ("-p", "--password"):
        P_WD = value
    elif opt in ("-l", "--logfile"):
        LOG_FILE = value
    elif opt in ("-n", "--port"):
        PORT_NUMBER = value
    elif opt in ("-h", "--help"):
        usage()
    else:
        usage()

if U_NAME is None or P_WD is None or LOG_FILE is None:
    usage()

HTTPServerRequestHandler.USERNAME = U_NAME
HTTPServerRequestHandler.PASSWD = P_WD

L_LOGGER_FILE = Logger("Profile-Push", LOG_FILE)
run(int(PORT_NUMBER))
