FROM jenkinsci/blueocean

USER root
RUN apk add py-pip
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
