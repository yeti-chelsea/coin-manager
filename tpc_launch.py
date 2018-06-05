#!/usr/bin/env python3

'''
TPC launcher
'''
import argparse
import signal
import sys
from log_manager import Logger
from udp_client_manager import UdpClientThread
from http_server_manager import HttpServer

LIST_OF_THREADS = []

def signal_handler(signalnum, stack):
    '''
    Signal handler for SIGINT, SIGQUIT, SIGHUP
    '''
    print(signalnum, stack)
    for thread_obj in LIST_OF_THREADS:
        thread_obj.stop()

def main():
    '''
    Main function
    '''

    parser = argparse.ArgumentParser(description='tpc launcher')
    parser.add_argument('-l', dest="logfile", \
	default="/var/log/tpc_launch.log", help='log file path')
    parser.add_argument('-u', dest='udp_server_address', \
        default="c-manager.rgowda.mycloud.wtl.sandvine.com:6767", \
		help='UDP server to connect')
    parser.add_argument('-t', dest='http_server', \
		default="0.0.0.0:6767", help="HTTP server to listen")

    args = parser.parse_args()
    log_file = str(args.logfile)
    udp_server_addr = str(args.udp_server_address).split(':')[0]
    udp_server_port = str(args.udp_server_address).split(':')[1]
    http_server_addr = str(args.http_server).split(':')[0]
    http_server_port = str(args.http_server).split(':')[1]

    l_logger = Logger("tpc_launch", log_file)

    l_logger.debug("Initaizing Signal Handlers")
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGQUIT, signal_handler)
    signal.signal(signal.SIGHUP, signal_handler)

    l_logger.info("Starting HTTP server thread")
    bind_address = (http_server_addr, int(http_server_port))
    httpserver = HttpServer(bind_address, l_logger)
    httpserver.start()
    LIST_OF_THREADS.append(httpserver)

    l_logger.info("Starting UDP client Thread")
    server_address = (udp_server_addr, int(udp_server_port))
    udpclient_manager = UdpClientThread(server_address, l_logger)
    udpclient_manager.start()
    LIST_OF_THREADS.append(udpclient_manager)

if __name__ == "__main__":
    main()
