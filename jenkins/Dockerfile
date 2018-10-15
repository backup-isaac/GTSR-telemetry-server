FROM jenkinsci/blueocean

USER root
RUN apk add py-pip
RUN pip install docker-compose 

COPY ./copy.docker /usr/local/bin/copy.docker	
RUN chmod +x /usr/local/bin/copy.docker	
RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/copy.docker" >> /etc/sudoers

RUN apk add sudo

RUN addgroup -S jenkins docker
RUN chown jenkins:docker /var/run/docker.sock
USER jenkins 
