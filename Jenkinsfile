pipeline {
  agent {
    docker {
      image 'golang:1.11.1'
      args '-w /go/src/telemetry-server -v ${PWD}/:/opt/telemetry-server'
    }

  }
  stages {
    stage('Build') {
      steps {
        sh 'cp -r /opt/telemetry-server/* /go/src/telemetry-server'
        sh 'cd /go/src/telemetry-server'
        sh 'go get -t ./...'
      }
    }
    stage('Test') {
      steps {
        sh 'go fmt ./...'
        sh 'go test ./...'
        sh 'go get golang.org/x/lint/golint'
        sh '../../bin/golint ./...'
      }
    }
  }
  environment {
    IN_DOCKER = 'true'
  }
}