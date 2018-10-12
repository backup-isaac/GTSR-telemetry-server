FROM jenkinsci/blueocean

USER root
RUN apk add py-pip
RUN pip install docker-compose 

RUN apk add sudo
COPY ./restart.docker /usr/local/bin/restart.docker
RUN chmod +x /usr/local/bin/restart.docker
RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/restart.docker" >> /etc/sudoers
USER jenkins 
