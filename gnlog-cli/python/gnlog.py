#!/usr/bin/python
# -*- coding:utf8 -*-

from logging import Handler
from Queue import Queue
from threading import Thread

LOG_BUF_SIZE = 100

def LogSender(queue):
    while True:
        msg = queue.get()
        #send msg
        #TODO
        queue.task_done()
        

class GNLogHandler(Handler):
    def __init__(self, strm):
        Handler.__init__(self)
        self.queue = Queue(LOG_BUF_SIZE)
        t = Thread(target=LogSender, args=(self.queue,))
        t.daemon = True
        t.start()

    def __del__(self):
        self.queue.join()

    def emit(self, record):
        try:
            msg = self.format(record)
            self.queue.put(msg)
        except:
            self.handleError()

def basicConfig():
    pass

def getLogger(name):
    gnhdlr = GNLogHandler()
    logger = logging.getLogger(name)
    logger.addHandler(GNLogHandler)
    return logger
