#!/bin/bash

#!/bin/sh
set -e

DEFAULT=help

#########################
## your functions here ##
#                       #

help() {
    echo "=================== help ========================"
    echo "help - run this help"
    echo "app - run app"
    echo "deploy - deploy app"
    echo "================================================="
}

setup_env() {
    go get
    go mod download
}

app() {
    go run main.go
}

build_with_ko() {
    # parameters:
    # $1 - additional tags, space separated

    local _kobuild_VERSION=0.15.1  # choose the latest version (without v prefix)
    local _kobuild_OS=Linux        # or Darwin
    local _kobuild_ARCH=x86_64     # or arm64, i386, s390x

    local _build_host="$(hostname)"
	local _build_date="$(date +"%Y-%m-%d_%H:%M:%S")"
    local _build_ldflags="-X main.BuildHost=${_build_host} -X main.BuildDate=${_build_date}"

    export APP_BUILD_HOST="${_build_host}"
    export APP_BUILD_DATE="${_build_date}"

    mkdir -p $(pwd)/bin/ko

    local _image_tags="latest"

    if [[ "x$1" != "x" ]]; then
        _image_tags="$1"
    fi

    echo "build with tags: $_image_tags"

    if [[ ! -f "$(pwd)/bin/ko/ko" ]]; then
        echo "ko.build not found, downloading it ... $(pwd)/bin/ko/ko.tar.gz"
        curl -sSfL "https://github.com/ko-build/ko/releases/download/v${_kobuild_VERSION}/ko_${_kobuild_VERSION}_${_kobuild_OS}_${_kobuild_ARCH}.tar.gz" > "$(pwd)/bin/ko/ko.tar.gz"
        
        # TODO: enable more security: https://ko.build/install/ 

        #curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/multiple.intoto.jsonl > multiple.intoto.jsonl
        #slsa-verifier verify-artifact --provenance-path multiple.intoto.jsonl --source-uri github.com/ko-build/ko --source-tag "v${VERSION}" ko.tar.gz 

        echo "unpack .."
        tar -vxzf "$(pwd)/bin/ko/ko.tar.gz" --directory "$(pwd)/bin/ko"
        ls -l "$(pwd)/bin/ko/"

        #echo "remove tar .."
        rm "$(pwd)/bin/ko/ko.tar.gz"

        echo "set rights .."
        chmod +x "$(pwd)/bin/ko/ko"
        echo "... done!"

    fi

    if [[ "x$DOCKER_REGISTRY" == "x" ]]; then
        export KO_DOCKER_REPO="ko.local"
    else
        export KO_DOCKER_REPO="$DOCKER_REGISTRY/$DOCKER_PATH"
        echo "$DOCKER_REGISTRY_PASS" | $(pwd)/bin/ko/ko login $DOCKER_REGISTRY --username $DOCKER_REGISTRY_USER --password-stdin
    fi;


    # {} gets replaced by piped name
    ./bin/ko/ko build \
        --tags="$_image_tags" \
        --image-label="org.opencontainers.image.title=download-upload-test" \
        --image-label="org.opencontainers.image.build-date=$APP_BUILD_DATE" \
        --image-label="org.opencontainers.image.build-host=$APP_BUILD_HOST" \
        --image-label="org.opencontainers.image.vendor=spitz.is" \
        --image-label="org.opencontainers.image.authors=spitz.is" \
        --sbom=none 
}

build() {
    mkdir -p ./build
    CGO_ENABLED=false go build -buildvcs=false -tags "osusergo netgo" -o "./build/download-upload-test"
    chmod +x ./build/download-upload-test
}


#                         #
## end of your functions ##
###########################

################################
## your helper functions here ##
#                              #


debug() {
    echo "debug: $1"
}


#                                #
## end of your helper functions ##
##################################

run() {
    if [ -e .env ]; then
        echo "found .env, sourcing it"
        . ./.env
    fi;

    # source run file
    . ./run.sh --source-only
    
    # run default function or run defined one
    if [ "x${1}" == "x" ]; then
        echo "run default: $DEFAULT"
        $DEFAULT
    else
        $@ "${@:2}"
    fi;
}

# this prevents running when sourcing
if [ "${1}" != "--source-only" ]; then
    run "${@}"
fi
