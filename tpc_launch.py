#!/usr/bin/env python3

from udp_client_manager import UdpClientThread 
from log_manager import Logger

def main():
    '''
    Main function
    '''
    l_logger = Logger("Ranjith", "stdout")
    server_address = ("localhost", 6767)
    udpclient_manager = UdpClientThread(server_address, l_logger)
    udpclient_manager.start()

if __name__ == "__main__":
    main()
