pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        sh 'cp -r . /go/src/telemetry-server'
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
        sh 'docker-compose build'
        sh '''docker-compose up -d --force-recreate influxdb
'''
        sh 'docker-compose up -d --force-recreate grafana'
        sh 'docker-compose up -d --force-recreate server'
        sh 'docker-compose restart nginx'
      }
    }
  }
}