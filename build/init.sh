#!/bin/bash

chown paas:paas /optemg/csplog/0/mediacache
chown paas:paas /opt/csplog/0/mediacache
chown paas:paas -R /opt/csp/mediacache
if ! grep -q "nameserver 8.8.8.8" /etc/resolv.conf; then
  echo "nameserver 8.8.8.8" >> /etc/resolv.conf
fi