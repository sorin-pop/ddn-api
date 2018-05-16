#!/bin/bash

if [[ $1 == "--push" ]]; then
    push="true"
    shift 1
fi

if [[ $# -ne 1 ]]; then
    echo 'Please specify the version. Should be major.minor.patch (e.g. 3.1.10).'
else
    version=$1
    # rootloc is the base root of the repository
    rootloc=`pwd`/..

    echo "building binary of server.."
    cd $rootloc
    date=`date -u "+%Y-%m-%d_%H:%M:%S"`
    commit=`git log -n 1 --pretty=format:"%H"`

    GOOS=linux go build -ldflags "-X main.version=${version} -X 'main.buildTime=`date`' -X main.commit=${commit}"
    
    cp $rootloc/ddn-api $rootloc/dist/ddn-api

    echo "updating libraries"
    cd $rootloc/web
    npm install -u

    echo "copying server.."
    cp -r $rootloc/web $rootloc/dist/web

    cd $rootloc/dist

    echo "building ddn-api image"
    docker build -t djavorszky/ddn-api:$version -t djavorszky/ddn-api:latest .

    echo "removing artefacts.."
    rm -rf $rootloc/dist/server/server $rootloc/dist/server/web

    if [[ $push == "true" ]]; then
        docker push djavorszky/ddn-api:$version
        docker push djavorszky/ddn-api:latest
    fi

    echo "cleanup"
    rm -rf $rootloc/dist/web $rootloc/dist/ddn-api
fi
