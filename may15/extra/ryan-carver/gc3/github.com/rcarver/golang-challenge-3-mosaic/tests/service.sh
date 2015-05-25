#!/bin/bash
# This script tests the mosaicly API to fetch images and generate a mosaic.
# It exits 0 on success. Any other exit is a failure.

set -e

# Write to stderr.
log() {
  echo "$@" 1>&2;
}

mosaicly=$GOPATH/bin/mosaicly
dir=`mktemp -d -t mosaicly`

port=8081
endpoint="localhost:$port"
tag="balloon"
img="balloon.jpg"

log "Starting server..."
$mosaicly serve -dir $dir -port $port -num 5 -units 10 &
sleep 1

log "Create a new mosaic..."
res="$(curl -fs -X POST -F img=@fixtures/${img} "$endpoint/mosaics?tag=${tag}")"
echo $res
id=$(echo $res | jq -M -r .id)
url=$(echo $res | jq -M -r .url)
img_url=$(echo $res | jq -M -r .img)
status=$(echo $res | jq -M -r .status)

# Show the URL
log "New mosaic: id:$id status:$status at ${url}"

# Wait for the image to be done.
while [[ "$status" != "created" ]]; do
  status=$(curl -fs "${endpoint}${url}" | jq -M -r .status)
  log "Waiting for mosaic..."
  sleep 1
done;

log "All mosaics..."
res=$(curl -fs "${endpoint}/mosaics")
echo $res | jsonpretty

log "Opening image..."
curl -fs ${endpoint}${img_url} > ${dir}/${img}
open ${dir}/${img}

killall mosaicly

