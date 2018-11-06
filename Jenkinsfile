pipeline {
  agent {
    docker {
      image 'golang:1.11.1'
      args '-w /go/src/telemetry-server -v ${PWD}/:/go/src/telemetry-server'
    }

  }
  stages {
    stage('Build') {
      steps {
        sh 'rsync -r . /go/src/telemetry-server --delete'
      }
    }
  }
  environment {
    IN_DOCKER = 'true'
  }
}