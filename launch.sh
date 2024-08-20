#!/bin/bash
set -e

trap 'killall kv-store' SIGINT

cd $(dirname $0)

killall kv-store || true
sleep 0.1

go install -v

kv-store --db-location=newyork.db --config-file=sharding.toml --shard=newyork -addr=localhost:8080 &
kv-store --db-location=newyork-replica.db --config-file=sharding.toml --shard=newyork -addr=localhost:8081 -replica &

kv-store --db-location=california.db --config-file=sharding.toml --shard=california -addr=127.0.0.2:8080 &
kv-store --db-location=california-replica.db --config-file=sharding.toml --shard=california -addr=127.0.0.22:8080 -replica &

kv-store --db-location=washington.db --config-file=sharding.toml --shard=washington -addr=127.0.0.3:8080 &
kv-store --db-location=washington-replica.db --config-file=sharding.toml --shard=washington -addr=127.0.0.33:8080 -replica &

kv-store --db-location=seattle.db --config-file=sharding.toml --shard=seattle -addr=127.0.0.4:8080 &
kv-store --db-location=seattle-replica.db --config-file=sharding.toml --shard=seattle -addr=127.0.0.44:8080 -replica &

wait