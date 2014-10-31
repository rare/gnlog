import gnlog
import logging
import time

host = "localhost"
port = 20000
auth = "abcdefg"
catalog = "www.123.com"
filename = "web.log"
fmt = "%(asctime)-15s %(levelname)-8s %(filename)s:%(lineno)s %(message)s"
if not gnlog.init(host, port, auth, catalog, filename, fmt):
	print("init gnlog error")	
	exit(1)

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.debug("debug message")
logger.info("info message")
logger.warning("warning message")
logger.error("error message")
logger.critical("critical message")

while True:
	time.sleep(1)

