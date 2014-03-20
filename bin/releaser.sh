#!/bin/bash

if [ "$1" = "release" ]
then
  # Init reqs
  clear

  echo ==============================
  echo "yapi Releaser"
  echo ==============================

  # Cleanup
  rm -rf shared/
  mkdir shared
  cd shared/

  # Get files
  wget -q https://raw.githubusercontent.com/cmfatih/yapi/master/pipe.json
  wget -q https://raw.githubusercontent.com/cmfatih/yapi/master/bin/releaser.sh
  chmod +x ./releaser.sh

  # Docker
  echo "Starting the docker container..."
  docker run -i -t -v `pwd`:/shared cmfatih/golang /shared/releaser.sh build

elif [ "$1" = "build" ]
then
  # Init reqs
  source /golang-crosscompile/crosscompile.bash

  echo "Building..."

  go get github.com/cmfatih/yapi
  cd src/github.com/cmfatih/yapi/
  go-build-all yapi.go
  mv yapi-* /shared
  cd /shared

  echo "Preparing binary files..."

  # Unix-like systems
  FILES="yapi-darwin-386 yapi-darwin-amd64 yapi-freebsd-386 yapi-freebsd-amd64 yapi-freebsd-arm yapi-linux-386 yapi-linux-amd64 yapi-linux-arm"

  for FILE in $FILES; do
    rm -rf yapi/
    mkdir yapi
    cp pipe.json yapi/pipe.json
    cp ${FILE} yapi/yapi
    rm -f ${FILE}.tar.gz
    tar -czf ${FILE}.tar.gz yapi/
    rm -f ${FILE}
    echo "Done! ${FILE}.tar.gz"
  done
  rm -rf yapi/

  # Windows
  FILES="yapi-windows-386 yapi-windows-amd64"

  for FILE in $FILES; do
    rm -rf yapi/
    mkdir yapi
    cp pipe.json yapi/pipe.json
    cp ${FILE} yapi/yapi.exe
    rm -f ${FILE}.zip
    zip -rq ${FILE}.zip yapi/
    rm -f ${FILE}
    echo "Done! ${FILE}.zip"
  done
  rm -rf yapi/

  # Cleanup
  rm pipe.json
  rm releaser.sh

  echo "Finished!"
else
  echo "Usage: ./releaser.sh release"
  echo 
fi
