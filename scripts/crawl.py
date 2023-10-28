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

def getPeersRequest(key):
  return '{{"keepalive":true, "request":"debug_remoteGetPeers", "arguments": {{"key":"{}"}}}}'.format(key)

def doRequest(req):
  try:
    ygg = socket.socket(socktype, socket.SOCK_STREAM)
    ygg.connect(sockaddr)
    ygg.send(req)
    data = json.loads(ygg.recv(1048576))
    return data
  except:
    return None

visited = set() # Add nodes after a successful lookup response
rumored = set() # Add rumors about nodes to ping
timedout = set()

def handleGetPeersResponse(publicKey, data):
  global vistied
  global rumored
  global timedout
  if publicKey in visited: return
  #visited.add(publicKey)
  try:
    ks = data['response'].values()[0]['keys']
    for k in ks:
      if k in visited: continue
      if k in timedout: continue
      rumored.add(k)
    visited.add(publicKey)
  except:
    pass

# Get self info
selfInfo = doRequest('{"keepalive":true, "request":"getSelf"}')
#rumored.add(selfInfo['response']['key'])
visited.add(selfInfo['response']['key'])
try:
  peers = doRequest('{"keepalive":true, "request":"getPeers"}')
  for p in peers['response']['peers']:
    rumored.add(p['key'])
except:
  pass

# Initialize dicts of visited/rumored nodes
#for k,v in selfInfo['response']['self'].iteritems(): rumored[k] = v

# Loop over rumored nodes and ping them, adding to visited if they respond
while len(rumored) > 0:
  for k in rumored:
    print "DEBUG tested:", len(visited), "remaining:", len(rumored)
    handleGetPeersResponse(k, doRequest(getPeersRequest(k)))
    break
  rumored.remove(k)
#End

# TODO do something with the results

#print visited
#print timedout
