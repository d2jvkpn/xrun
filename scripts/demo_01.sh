#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

make build

./xrun version --json

./xrun run -y examples/pipeline.yaml -t sleep -p 0
ls temp/
