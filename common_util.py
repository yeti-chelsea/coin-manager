'''
This file contains all the common functions
'''

import os
import json
import sys
from subprocess import PIPE, Popen

MINER_DAEMON_PATH = "/opt/cypto"
MINER_PROCESS_NAME = "svsdem"
MINER_LOG_FILE = "/var/tmp/m-svsde.log"
MINER_DAEMON_CC = MINER_DAEMON_PATH + "/cc/bin/svsdem"
MINER_DAEMON_XMR = MINER_DAEMON_PATH + "/xmr/bin/svsdem"
MINER_DAEMON_IPBC = MINER_DAEMON_PATH + "/ipbc/bin/svsdem"
MINER_DAEMON_WEBCHAIN = MINER_DAEMON_PATH + "/webchain/bin/svsdem"

CONFIG_CC_PATH = MINER_DAEMON_PATH + "/cc/etc/"
CONFIG_XMR_PATH = MINER_DAEMON_PATH + "/xmr/etc/"
CONFIG_IPBC_PATH = MINER_DAEMON_PATH + "/ipbc/etc/"
CONFIG_WEBCHAIN_PATH = MINER_DAEMON_PATH + "/webchain/etc/"

def cmdline(command):
    '''
    A simple method to execute all shell commands
    '''
    process = Popen(
        args=command,
        stdout=PIPE,
        shell=True
    )
    return process.communicate()[0]

def get_miner_daemons():
    '''
    Method to get all the miner daemons that can be
    deployed
    '''
    return next(os.walk(MINER_DAEMON_PATH))[1]

def get_miner_coins():
    '''
    Method to get all the coins that are supported by
    various miner daemons
    '''
    list_of_minerd = get_miner_daemons()

    coins = {}
    for each_miner in list_of_minerd:
        miner_etc = MINER_DAEMON_PATH + "/" + each_miner \
                + "/" + "etc/"
        coins[each_miner] = next(os.walk(miner_etc))[2]

    list_of_coins = []
    for _, values in coins.items():
        for value in values:
            if value == "any_config.txt" or value == "any_cpu.txt":
                continue

            list_of_coins.append(value.split("_")[0])

    return list_of_coins

def get_miner_coin_and_daemon():
    '''
    Method to get both miner daemon and its
    associated coins
    '''
    list_of_minerd = get_miner_daemons()

    daemon_coins = {}
    for each_miner in list_of_minerd:
        miner_etc = MINER_DAEMON_PATH + "/" + each_miner \
                + "/" + "etc/"
        daemon_coins[each_miner] = next(os.walk(miner_etc))[2]

    final_dict = {}

    for key, values in daemon_coins.items():
        list_of_coins = []
        for value in values:
            if value == "any_config.txt" or value == "any_cpu.txt":
                continue
            list_of_coins.append(value.split("_")[0])
        final_dict[key] = list_of_coins

    return final_dict
def get_miner_daemons_json():
    '''
    Get the miner daemons in json format
    {
        "miner-daemons": "list of daemons"
    }
    '''
    data = {}
    data['miner-daemons'] = get_miner_daemons()
    return json.dumps(data)

def get_miner_coins_json():
    '''
    Get miner coins in json format
    {
        "miner-coins": "list of coins"
    }
    '''
    data = {}
    data['miner-coins'] = get_miner_coins()
    return json.dumps(data)

def check_miner_running_status():
    '''
    Find whether miner process is running and returns
    the command line args provided
    '''
    svsdem_pid = cmdline('pidof svsdem').decode().rstrip("\n")

    if len(svsdem_pid) < 1:
        return ""

    svsdem_cmdline_options = "xargs -0 < /proc/" + svsdem_pid + "/cmdline"
    return cmdline(svsdem_cmdline_options).decode().rstrip("\n")

def get_current_coin():
    '''
    Get current coin which is mining
    '''
    mine_status = check_miner_running_status()
    if len(mine_status) < 1:
        return "Miner daemon not running"

    return mine_status


def stop_mining():
    '''
    method to stop mining
    '''

    if len(check_miner_running_status()) < 1:
        return "Miner daemon not running"

    command_to_stop_mining = 'kill -9 pidof ' + MINER_PROCESS_NAME
    cmdline(command_to_stop_mining)
    return "Success"

def get_mine_log():
    '''
    method to get miner logs
    '''
    if os.path.isfile(MINER_LOG_FILE):
        strings_list = ['' for i in range(10)]

        with open(MINER_LOG_FILE, 'r') as f:
            for line in f:
                strings_list.pop(0)
                strings_list.append(line)

        return ''.join(strings_list)
    else:
        return "Miner daemon not running"

def start_mining(mine_coin):
    '''
    Method to start mining
    '''

    if len(check_miner_running_status()) < 1:
        return "Miner daemon not running"

    miner_daemon = ""
    m_daemon_coin = get_miner_coin_and_daemon()

    mine_coin_found = False
    for key, values in m_daemon_coin.items():
        for coin in values:
            if coin == mine_coin:
                miner_daemon = key
                mine_coin_found = True
                break
        if mine_coin_found == True:
            break

    if mine_coin_found == False:
        return "Coin not supported"

    miner_daemon_path = ""
    if miner_daemon == 'xmr':
        miner_daemon_path = MINER_DAEMON_XMR
    elif miner_daemon == 'cc':
        miner_daemon_path = MINER_DAEMON_CC
    elif miner_daemon == 'ipbc':
        miner_daemon_path = MINER_DAEMON_IPBC
    elif miner_daemon == 'webchain':
        miner_daemon_path = MINER_DAEMON_WEBCHAIN
    else:
        return "No miner daemon present for the coin"

    daemon_cmd_line_option = ""
    if miner_daemon == 'cc':
        config_file = CONFIG_CC_PATH + mine_coin + '_config.json'
        daemon_cmd_line_option = ' -c' + CONFIG_CC_PATH + config_file
    elif miner_daemon == 'xmr':
        config_file = CONFIG_XMR_PATH + mine_coin + '_pool.txt'
        any_config = CONFIG_XMR_PATH + 'any_config.txt'
        any_cpu = CONFIG_XMR_PATH + 'any_cpu.txt'
        daemon_cmd_line_option = ' -c ' + any_config + ' -C ' + config_file + ' --cpu ' + any_cpu
    elif miner_daemon == 'ipbc':
        config_file = CONFIG_IPBC_PATH + 'ipbc_pool.txt'
        any_config = CONFIG_IPBC_PATH + 'any_config.txt'
        any_cpu = CONFIG_IPBC_PATH + 'any_cpu.txt'
        daemon_cmd_line_option = ' -c ' + any_config + ' -C ' + config_file + ' --cpu ' + any_cpu
    elif miner_daemon == 'webchain':
        config_file = CONFIG_IPBC_PATH + 'webchain_pool.txt'
        any_config = CONFIG_WEBCHAIN_PATH + 'any_config.txt'
        any_cpu = CONFIG_WEBCHAIN_PATH + 'any_cpu.txt'
        daemon_cmd_line_option = ' -c ' + any_config + ' -C ' + config_file + ' --cpu ' + any_cpu
    else:
        return "Miner config not found"

    final_cmd = miner_daemon_path + daemon_cmd_line_option + ' >/dev/null 2>&1 &'
    print(final_cmd)
    cmdline(final_cmd)
    return "Success"

