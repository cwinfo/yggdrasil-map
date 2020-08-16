#!/usr/bin/env sh

set -e

cron
#cd /src/yggdrasil-map/scripts/ && python crawl-dht.py
#cd /src/yggdrasil-map/web/ && python updateGraph.py
cd /src/yggdrasil-map/web/static/ && wget http://[21f:dd73:7cdb:773b:a924:7ec0:800b:221e]/static/graph.json
python /src/yggdrasil-map/web/web.py --host $HOST --port $PORT
exit $?