#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)/
_path=$(dirname $0 | xargs -i readlink -f {})

echo a2
sleep 10

