FROM jenkinsci/blueocean
USER root
# Installs golang 1.11.2 from Dockerfile here
# https://github.com/docker-library/golang/blob/master/1.11/alpine3.8/Dockerfile
RUN apk add --no-cache \
		ca-certificates

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

RUN apk add py-pip
RUN apk add rsync
RUN pip install docker-compose 

COPY ./copy.docker /usr/local/bin/copy.docker	
RUN chmod +x /usr/local/bin/copy.docker	


COPY ./change-socket.docker /usr/local/bin/change-socket.docker	
RUN chmod +x /usr/local/bin/change-socket.docker	


RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/copy.docker" >> /etc/sudoers
RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/change-socket.docker" >> /etc/sudoers


RUN apk add sudo
RUN addgroup -S jenkins docker
USER jenkins 
