pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        sh 'rsync -r . /go/src/telemetry-server --delete'
        sh 'cd $GOPATH/src/telemetry-server && go get -v -t ./...'
      }
    }
    stage('Test') {
      steps {
        sh 'cd $GOPATH/src/telemetry-server && go fmt ./...'
        sh 'cd $GOPATH/src/telemetry-server && go test ./...'
        sh 'cd $GOPATH/src/telemetry-server && go get golang.org/x/lint/golint'
        sh 'cd $GOPATH/src/telemetry-server && ../../bin/golint ./...'
      }
    }
    stage('Deploy') {
      steps {
        sh 'sudo change-socket.docker'
        sh 'sudo copy.docker'
        sh 'cd /opt/telemetry-server && docker-compose build'
        sh 'cd /opt/telemetry-server && docker-compose up -d --force-recreate influxdb'
        sh 'cd /opt/telemetry-server && docker-compose up -d --force-recreate grafana'
        sh 'cd /opt/telemetry-server && docker-compose up -d --force-recreate server'
        sh 'cd /opt/telemetry-server && docker-compose restart nginx'
      }
    }
  }
}