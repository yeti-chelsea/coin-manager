#!/usr/bin/env python3

from udp_client_manager import UdpClientThread 
from log_manager import Logger

def main():
    '''
    Main function
    '''
    l_logger = Logger("tpc_launch", "/var/log/tpc_launch.log")
    server_address = ("c-manager.rgowda.mycloud.wtl.sandvine.com", 6767)
    udpclient_manager = UdpClientThread(server_address, l_logger)
    udpclient_manager.start()

if __name__ == "__main__":
    main()
