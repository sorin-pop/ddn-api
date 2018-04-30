#!/bin/bash

# rootloc is the base root of the repository
rootloc=`pwd`/../..

echo "Preparing build container"
cp -r $rootloc/agent $rootloc/server $rootloc/common .


docker build -f Dockerfile.prepare -t djavorszky/ddn:build .

docker rm builder

docker container create --name builder djavorszky/ddn:build  
docker commit builder
docker push djavorszky/ddn:build

rm -rf agent server common
echo "done"
