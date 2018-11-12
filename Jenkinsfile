pipeline {
  agent none
  stages {
    stage('Build') {
        agent {
            docker {
                image 'golang:1.11.2'
                args '--mount source=go-cache,target=/go'
            }
        }
        steps {
            slackSend color: "#439FE0", message: "Build Started: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
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
      }
    }
    stage('Deploy') {
      agent any
      when {
        branch 'master'
      }
      steps {
        sh 'go fmt ./...'
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
  post {
       // only triggered when blue or green sign
       success {
           slackSend color: "good", message: "Job Succeeded: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
       }
       // triggered when red sign
       failure {
           slackSend color: "danger", message: "Job Failed: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
       }
  }
}
