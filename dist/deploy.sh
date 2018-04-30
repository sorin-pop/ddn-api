#!/bin/bash

if [[ $# -ne 1 ]]; then
    echo 'Please specify the version. Should be major.minor.patch (e.g. 3.1.10).'
else
    version=$1
    # rootloc is the base root of the repository
    rootloc=`pwd`/../..

    cp -r $rootloc/agent $rootloc/server $rootloc/common .

    echo "building binary of server.."
    docker build -f Dockerfile.build -t djavorszky/ddn:build .

    rm -rf $rootloc/dist/server/agent $rootloc/dist/server/server $rootloc/dist/server/common

    docker container create --name extract djavorszky/ddn:build  
    docker container cp extract:/go/src/github.com/djavorszky/ddn/server/server $rootloc/dist/server/server
    docker container rm -f extract

    echo "updating libraries"
    cd $rootloc/server/web
    npm install -u

    echo "copying server.."
    cp -r $rootloc/server/web $rootloc/dist/server/web

    cd $rootloc/dist/server

    echo "building server image"
    docker build -t djavorszky/ddn:$version -t djavorszky/ddn:latest .

    echo "stopping previous version"
    docker stop ddn-server
    docker rm ddn-server

    echo "starting container.."
    docker run -dit -p 7010:7010 --name ddn-server -v $rootloc/dist/data:/ddn/data -v $rootloc/dist/ftp:/ddn/ftp djavorszky/ddn:$version

    echo "removing artefacts.."
    rm -rf $rootloc/dist/server/server $rootloc/dist/server/web

    docker push djavorszky/ddn:$version
    docker push djavorszky/ddn:latest
fi
