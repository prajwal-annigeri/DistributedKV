# DistributedKV

DistributedKV is a distributed key-value store with support for replication and resharding. Built with Go using Bolt.

# Getting Started
## Requirements
* `Go` >= 1.21.4

## Fetch the application
`go get github.com/prajwal-annigeri/kv-store`

## Modify `sharding.toml`
Each shard needs 4 fields:
* name(string): name of the shard
* idx(int): shard index
* address(string): address for that particular shard
* replica-address(list[string]): list of addresses for the replicas of this particular shard

You can add or remove shards by  adding/removing a shard object from the `sharding.toml` file.

## Modify `launch.sh`
### Options for running an instance
| Flag | Value |
| :--- | --- |
| --db-location |  Database file |
| --config-file | shard config filename|
| -shard | name of the shard|
| -addr | address for this instance |
| -replica| boolean: use this option when startin a replica instance|

Please refer to launch.sh which currently has specified 8 instances (including 4 replicas). You can add or remove instances

# Usage
## Running the application
Run the `launch.sh` script

## To store a key-value pair
Make a HTTP GET rquest to the /set endpoint

`curl http://{address_of_instance}/set?key={key_here}&value={value_here}`

## To fetch value for a key
Make a HTTP GET rquest to the /get endpoint

`curl http://{address_of_instance}/get?key={key_here}`