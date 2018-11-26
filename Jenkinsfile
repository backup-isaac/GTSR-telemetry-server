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
                args '--mount source=go-cache,target=/go'
            }
        }
        steps {
            sh 'rm -r -f $GOPATH/src/telemetry-server'
            sh 'mkdir -p $GOPATH/src/telemetry-server'
            sh 'cp -r . $GOPATH/src/telemetry-server'
            sh 'cd $GOPATH/src/telemetry-server && go get -v -t ./...'
        }
    }
    stage('Test') {
      agent {
        docker {
                image 'golang:1.11.2'
                args '--mount source=go-cache,target=/go'
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
                args '--mount source=go-cache,target=/go'
            }
      }
      steps {
        sh 'cd $GOPATH/src/telemetry-server && go get golang.org/x/lint/golint'
        sh 'cd $GOPATH/src/telemetry-server && ../../bin/golint ./...'
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
        sh 'go fmt ./...'
        sh 'sudo change-socket.docker'
        sh 'sudo copy.docker'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml build'
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
