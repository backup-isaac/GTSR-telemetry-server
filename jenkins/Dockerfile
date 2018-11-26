FROM jenkinsci/blueocean
USER root

# Installs docker-compose and rsync
RUN apk add py-pip
RUN pip install docker-compose 
RUN apk add rsync
RUN apk add sudo

# Installs our copy and change-socket scripts
# and grants the jenkins user sudo access to them
COPY ./copy.docker /usr/local/bin/copy.docker	
RUN chmod +x /usr/local/bin/copy.docker	

COPY ./change-socket.docker /usr/local/bin/change-socket.docker	
RUN chmod +x /usr/local/bin/change-socket.docker	

RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/copy.docker" >> /etc/sudoers
RUN echo "jenkins  ALL = (root) NOPASSWD: /usr/local/bin/change-socket.docker" >> /etc/sudoers

RUN addgroup -S jenkins docker
USER jenkins 
