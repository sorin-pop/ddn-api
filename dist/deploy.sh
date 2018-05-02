#!/bin/bash

if [[ $# -ne 1 ]]; then
    echo 'Please specify the version. Should be major.minor.patch (e.g. 3.1.10).'
else
    version=$1
    # rootloc is the base root of the repository
    rootloc=`pwd`/..

    echo "building binary of server.."
    cd $rootloc
    GOOS=linux go build -ldflags "-X main.version=`date -u +%Y%m%d.%H%M%S`"

    cp $rootloc/ddn-api $rootloc/dist/ddn-api

    echo "updating libraries"
    cd $rootloc/web
    npm install -u

    echo "copying server.."
    cp -r $rootloc/web $rootloc/dist/web

    cd $rootloc/dist

    echo "building ddn-api image"
    docker build -t djavorszky/ddn-api:$version -t djavorszky/ddn-api:latest .

    echo "stopping previous version"
    docker stop ddn-api
    docker rm ddn-api

    echo "removing artefacts.."
    rm -rf $rootloc/dist/server/server $rootloc/dist/server/web

    docker push djavorszky/ddn-api:$version
    docker push djavorszky/ddn-api:latest

    #echo "starting container.."
    #docker run -dit -p 7010:7010 --name ddn-server -v $rootloc/dist/data:/ddn/data -v $rootloc/dist/ftp:/ddn/ftp djavorszky/ddn:$version

fi
