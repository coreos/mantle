#!groovy

properties([
    buildDiscarder(logRotator(daysToKeepStr: '20', numToKeepStr: '30')),

    [$class: 'GithubProjectProperty',
     projectUrlStr: 'https://github.com/coreos/mantle'],

    [$class: 'CopyArtifactPermissionProperty',
     projectNames: '*'],

    parameters([
        choice(name: 'GOARCH',
               choices: "amd64\narm64",
               description: 'target architecture for building binaries')
    ]),

    pipelineTriggers([pollSCM('H/15 * * * *')])
])

node('docker') {
    stage('SCM') {
        sh "echo testing"
    }
}
