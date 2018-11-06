pipeline {
  agent {
    docker {
      image 'golang:1.11.1'
      args '-v ${PWD}/:/opt/telemetry-server'
    }

  }
  stages {
    stage('Build') {
      steps {
        sh 'cp -r . /go/src/telemetry-server'
        sh 'cd /go/src/telemetry-server'
        sh 'ls'
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
  }
}