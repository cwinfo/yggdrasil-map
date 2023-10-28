import json
import socket
import sys
import time

#gives the option to get data from an external server instead and send that
#if no options given it will default to localhost instead
if len(sys.argv) == 3:
  socktype = socket.AF_INET
  sockaddr = (sys.argv[1], int(sys.argv[2]))
elif len(sys.argv) == 2:
  socktype = socket.AF_UNIX
  sockaddr = sys.argv[1]
else:
  socktype = socket.AF_UNIX
  sockaddr = "/var/run/yggdrasil.sock"

def doRequest(req):
  try:
    ygg = socket.socket(socktype, socket.SOCK_STREAM)
    ygg.connect(sockaddr)
    ygg.send(req)
    data = json.loads(ygg.recv(1048576))
    return data
  except:
    return None

known = doRequest('{"keepalive":true, "request":"lookups"}')

infos = dict()
for node in known['response']['infos']:
  coords = json.dumps(node["path"]).replace(",", " ")
  v = {"address": node["addr"], "coords": coords, "time": node["time"]}
  infos[node["key"]] = v

# Initialize dicts of visited/rumored nodes
#for k,v in selfInfo['response']['self'].iteritems(): rumored[k] = v

# Loop over rumored nodes and ping them, adding to visited if they respond
print json.dumps({"yggnodes": infos})

# TODO do something with the results

#print visited
#print timedout
