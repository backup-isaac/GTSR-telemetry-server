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
        sh 'cp -r telemetry-server/* /go/src/telemetry-server'
      }
    }
  }
  environment {
    IN_DOCKER = 'true'
  }
}