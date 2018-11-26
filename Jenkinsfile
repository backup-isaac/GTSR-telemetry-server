pipeline {
  agent none
  stages {
    stage('Pre-build') {
        agent any
        steps {
            slackSend color: "#439FE0", message: "Build Started: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
            // necessarry to grant jenkins the mounted docker socket
            // in order to spin up more docker containers below
            sh 'sudo change-socket.docker' 
        }
    }
    stage('Build') {
        agent {
            docker {
                image 'golang:1.11.2'
                args '-v go-cache:/go'
                reuseNode true
            }
        }
        steps {
            // remove old builds
            sh 'rm -r -f /go/src/telemetry-server'
            // create src directory if it doesn't exist
            sh 'mkdir -p $GOPATH/src'
            // symlink current directory into gopath
            sh 'ln -s $WORKSPACE $GOPATH/src/telemetry-server'
            sh 'cd $GOPATH/src/telemetry-server && go get -t ./...'
        }
    }
    stage('Test') {
      agent {
        docker {
                image 'golang:1.11.2'
                args "-v go-cache:/go"
                reuseNode true
            }
      }
      steps {
        sh 'cd $GOPATH/src/telemetry-server && go fmt ./...'
        sh 'cd $GOPATH/src/telemetry-server && go test ./...'
      }
    }
    stage('Lint') {
      agent {
            docker {
                image 'golang:1.11.2'
                args '-v go-cache:/go'
                reuseNode true
            }
      }
      steps {
        sh 'cd $GOPATH/src/telemetry-server && go get golang.org/x/lint/golint'
        sh 'cd $GOPATH/src/telemetry-server && $GOPATH/bin/golint ./...'
        slackSend color: "#439FE0", message: "Build Suceeded: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
      }
    }
    stage('Deploy') {
      agent any
      when {
        branch 'master'
      }
      steps {
        slackSend color: "#439FE0", message: "Deploy Started: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
        sh 'sudo change-socket.docker'
        sh 'sudo copy.docker'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml build --pull'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml up -d --force-recreate influxdb'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml up -d --force-recreate grafana'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml up -d --force-recreate server'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml restart nginx'
        slackSend color: "good", message: "Deploy Succeeded: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
      }
    }
  }
  post {
       // triggered when red sign
       failure {
           slackSend color: "danger", message: "Job Failed: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
       }
  }
}
