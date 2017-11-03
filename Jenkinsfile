#!groovy

properties([
    buildDiscarder(logRotator(daysToKeepStr: '20', numToKeepStr: '30')),

    [$class: 'CopyArtifactPermissionProperty',
     projectNames: '*'],

    pipelineTriggers([pollSCM('H/15 * * * *')])
])

def docker_args = "-e CGO_ENABLED=1 --rm -u \"\$(id -u):\$(id -g)\" \
        -v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro -v \"\$PWD\":/usr/src/myapp \
        -w /usr/src/myapp mantle-builder:1.9.1.0"

def buildMantle = { String targetArch, String targetCC, String extraArgs ->
    def args = "-e GOARCH=${targetArch} -e CC=${targetCC} ${extraArgs} ${docker_args}"
    sh "docker run ${args} ./build"
}

def build_map = [:]

build_map.failFast = false
build_map['amd64'] = buildMantle.curry('amd64', 'gcc', '')
build_map['arm64'] = buildMantle.curry('arm64', 'aarch64-linux-gnu-gcc', '-e NO_CROSS=1')

node('amd64 && docker') {
    stage('SCM') {
        checkout scm
    }

    stage('Build Docker') {
        sh "bash -x ./build-docker --force"
    }

    stage('Build') {
        build_map.each {
            if (it.value)
                it.value.call()
        }
    }

    stage('Test') {
        sh "docker run ${docker_args} ./test -v"
    }

    stage('Post-build') {
        if (env.JOB_BASE_NAME == "master-builder") {
            archiveArtifacts artifacts: 'bin/**', fingerprint: true, onlyIfSuccessful: true
        }
    }
}
