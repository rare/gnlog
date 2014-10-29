#!/usr/bin/python
# -*- coding:utf8 -*-

from logging import Handler
from Queue import Queue
from threading import Thread

LOG_BUF_SIZE = 100

def LogSender(queue, host, port, auth, catalog, filename):
    while True:
        msg = queue.get()
        #send msg
        #TODO
        queue.task_done()
        

class GNLogHandler(Handler):
    def __init__(self, host, port, auth, catalog, filename):
        Handler.__init__(self)
		self.queue = Queue(LOG_BUF_SIZE)
		self.host = host
		self.port = port
		self.auth = auth
		self.catalog = catalog
		self.filename = filename
        t = Thread(target=LogSender, args=(self.queue, self.host, self.port, self.auth, self.catalog, self.filename))
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
