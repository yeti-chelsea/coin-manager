#!/usr/bin/env python3

import argparse
from udp_client_manager import UdpClientThread 
from log_manager import Logger

def main():
    '''
    Main function
    '''

    parser = argparse.ArgumentParser(description='tpc launcher')
    parser.add_argument('-l', dest="logfile", default="/var/log/tpc_launch.log", help='log file path')
    parser.add_argument('-u', dest='udp_server_address', \
            default="c-manager.rgowda.mycloud.wtl.sandvine.com:6767", help='UDP \
            server address')
    parser.add_argument('-t', dest='http_server', \
            default="0.0.0.0:6767", help="HTTP server to listen")
    
    args = parser.parse_args()
    log_file = str(args.logfile)
    udp_server_addr = str(args.udp_server_address).split(':')[0]
    udp_server_port = str(args.udp_server_address).split(':')[1]
    http_server_addr = str(args.http_server).split(':')[0]
    http_server_port = str(args.http_server).split(':')[1]

    print(udp_server_addr, udp_server_port)
    print(http_server_addr, http_server_port)

    l_logger = Logger("tpc_launch", log_file)
    server_address = (udp_server_addr, int(udp_server_port))
    udpclient_manager = UdpClientThread(server_address, l_logger)
    udpclient_manager.start()

if __name__ == "__main__":
    main()
