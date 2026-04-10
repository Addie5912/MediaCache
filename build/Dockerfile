FROM ubuntu:latest
LABEL authors="Administrator"

USER root
RUN mkdir -p /opt/csp/mediacache/module
COPY . /opt/csp/mediacache/module
RUN cat /opt/csp/mediacache/module/sudoer.config >> /etc/sudoers && \
    rm -rf /opt/csp/mediacache/module/sudoer.config && \
    rm -rf /opt/csp/mediacache/module/Dockerfile && \
    chown paas:paas -R /opt/csp/mediacache && \
    chmod +x /opt/csp/mediacache/module/mediacache && \
    chmod +x /opt/csp/mediacache/module/*.sh \

USER paas


ENTRYPOINT ["top", "-b"]