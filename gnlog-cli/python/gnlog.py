#!/usr/bin/python
# -*- coding:utf8 -*-

from threading import Thread
from Queue import Queue
from socket import socket, error

import logging
import struct
import json

LOG_BUF_SIZE = 100
SOCK_TIMEOUT = 3.0

HEADER_SIZE	 = 12
MAX_BODY_LEN = 4096

CMD_START	 = 1
CMD_LOG		 = 2

VERSION		 = 0

MODE_LINE	 = 1
MODE_STRING	 = 2

STATUS_OK	 = 0
STATUS_AUTH_ERR = -1

g_seq = 0

def _pack_start_cmd(auth):
	global g_seq
	sc = {
		"auth": auth,
		"mode": MODE_LINE
	}
	body = json.dumps(sc)
	header = struct.pack("!HHII", CMD_START, VERSION, g_seq, len(body))
	g_seq += 1
	return header + body

def _pack_log_cmd(catalog, filename, msg):
	global g_seq
	lc = {
		"catalog": catalog,
		"filename": filename,
		"logdata": msg
	}
	body = json.dumps(lc)
	header = struct.pack("!HHII", CMD_LOG, VERSION, g_seq, len(body))
	g_seq += 1
	return header + body

def _start_channel(host, port, auth):
	global g_seq
	s = socket()
	s.settimeout(SOCK_TIMEOUT)
	try: 
		s.connect((host, port))
		buf = _pack_start_cmd(auth)
		s.send(buf)

		resp_header = s.recv(HEADER_SIZE)
		cmd, ver, g_seq, len = struct.unpack('!HHII', resp_header)
		if cmd == CMD_START and len < MAX_BODY_LEN:
			pass
		else:
			print("invalid response header, cmd: %d, len: %d" % cmd, len)
			return None
		resp_body = s.recv(len)
		sr = json.loads(resp_body)
		if sr["status"] != STATUS_OK:
			print("invalid response status: %d" % sr.status)
			return None

	except error as e:
		print(e)
		return None
	return s

def _send_msg(sock, catalog, filename, msg):
	buf = _pack_log_cmd(catalog, filename, msg)
	try:
		sock.send(buf)
	except error as e:
		print(e)

def LogSender(queue, sock, catalog, filename):
    while True:
		msg = queue.get()
		_send_msg(sock, catalog, filename, msg)
		queue.task_done()
        
class GNLogHandler(logging.Handler):
	queue = None

	def __init__(self, sock, catalog, filename):
		logging.Handler.__init__(self)
		self.queue = Queue(LOG_BUF_SIZE)
		t = Thread(target=LogSender, args=(self.queue, sock, catalog, filename))
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


def init(host, port, auth, catalog, filename, fmt):
	sock = _start_channel(host, port, auth)
	if sock == None:
		return False
	handler = GNLogHandler(sock, catalog, filename)
	handler.setFormatter(logging.Formatter(fmt))
	logger = logging.getLogger()
	logger.handlers = [handler]
	return True

