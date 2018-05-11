
'''
A wrapper class for logging
'''

import logging
from logging.handlers import RotatingFileHandler

class Logger(object):
    '''
    A wrapper class for logging
    '''
    # pylint: disable=redefined-variable-type
    def __init__(self, logger_name, file_name):

        # create logger
        self._logger = logging.getLogger(logger_name)
        self._logger.setLevel(logging.DEBUG)

        # Create File Handler
        if file_name == "stdout":
            file_handler = logging.StreamHandler()
        else:
            file_handler = RotatingFileHandler(file_name, maxBytes=1024*1024*10, backupCount=10)

        file_handler.setLevel(logging.DEBUG)

        # create formatter
        formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')

        # add formatter to file handler
        file_handler.setFormatter(formatter)

        # add file handler to logger
        self._logger.addHandler(file_handler)

    def debug(self, f_arg, *args):
        '''
        Log Debug messages
        '''

        self._logger.debug(f_arg + " ".join(map(str, args)))

    def info(self, f_arg, *args):
        '''
        Log info messages
        '''
        self._logger.info(f_arg + " ".join(map(str, args)))

    def warning(self, f_arg, *args):
        '''
        Log warning messages
        '''
        self._logger.warning(f_arg + " ".join(map(str, args)))

    def error(self, f_arg, *args):
        '''
        Log error messages
        '''
        self._logger.error(f_arg + " ".join(map(str, args)))

    def critical(self, f_arg, *args):
        '''
        Log critical message
        '''
        self._logger.critical(f_arg + " ".join(map(str, args)))

    def change_loglevel(self, level_num):
        '''
        Change the log level
        0 - Disable logging
        10 - Debug logs
        20 - Info logs
        30 - Warning logs
        40 - Error logs
        50 - Critical logs
        '''
        self._logger.setLevel(level_num)
