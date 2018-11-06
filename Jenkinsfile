pipeline {
  agent {
    docker {
      image '1.11-alpine-3.8'
    }

  }
  stages {
    stage('Build') {
      steps {
        sh 'go build'
      }
    }
  }
}