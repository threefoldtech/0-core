import logging
import os

# Start testsuite logger
logger = logging.getLogger('zos_testsuite')
if not os.path.exists('logs/zos_testsuite.log'):
    os.mkdir('logs')
handler = logging.FileHandler('logs/zos_testsuite.log')
formatter = logging.Formatter('%(asctime)s [%(testid)s] [%(levelname)s] %(message)s',
                              '%d-%m-%Y %H:%M:%S %Z')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)
