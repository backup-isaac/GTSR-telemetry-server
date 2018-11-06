pipeline {
  agent {
    docker {
      image 'golang:1.11.1'
      args '''-w /go/src/telemetry-server 
-v ./:/go/src/telemetry-server
'''
    }

  }
  stages {
    stage('Build') {
      steps {
        sh '''go get -v -t ./... && 
go run main.go'''
      }
    }
  }
  environment {
    IN_DOCKER = 'true'
  }
}