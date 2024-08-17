#!/bin/bash
set -e

trap 'killall kv-store' SIGINT

cd $(dirname $0)

killall kv-store || true
sleep 0.1

go install -v

kv-store --db-location=newyork.db --config-file=sharding.toml --shard=newyork --web-port=8080 &
kv-store --db-location=california.db --config-file=sharding.toml --shard=california  --web-port=8081 &
kv-store --db-location=washington.db --config-file=sharding.toml --shard=washington  --web-port=8082 &

wait