#!/bin/bash -x

version=$(git rev-parse HEAD)

SHA=${version:0:8}

# output version file
echo $SHA > $WERCKER_ROOT/VERSION

# output archive name, that will be used for version archive name
echo `date "+%Y-%m-%dT%H:%M:%S"`_$SHA.zip > $WERCKER_ROOT/ARCHIVE_NAME
