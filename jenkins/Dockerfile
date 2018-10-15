FROM jenkinsci/blueocean

USER root
RUN apk add py-pip
RUN pip install docker-compose 

RUN apk add sudo
USER jenkins 
