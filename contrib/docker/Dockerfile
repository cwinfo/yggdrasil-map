FROM debian:buster
LABEL maintainer="Christer Waren/CWINFO <christer.waren@cwinfo.org>"

ENV HOST localhost
ENV PORT 3000

WORKDIR /src

RUN apt-get update && \
    apt-get install -y git wget python-flask python-flup python-mysqldb python-pygraphviz python-networkx cron && \
    apt-get clean

RUN git clone -b crawler https://github.com/Arceliar/yggdrasil-map.git yggdrasil-map

COPY web_config.cfg /src/yggdrasil-map/web/web_config.cfg
#COPY updateGraph.py /src/yggdrasil-map/web/updateGraph.py

COPY entrypoint.sh /usr/bin/entrypoint.sh
RUN chmod 0555 /usr/bin/entrypoint.sh

COPY crontab /etc/cron.d/jobs
RUN chmod 0644 /etc/cron.d/jobs

ENTRYPOINT [ "/usr/bin/entrypoint.sh"]