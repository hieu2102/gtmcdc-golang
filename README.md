## GT.M/YottaDB Replication Filter

A GT.M/YottaDB change event capture (CDC) mechansm by using a replication filter that convert M instructions into Kafka messages for the purpose of propagating change event to external systems.

The code here is just a skeleton. A lot more business logic needs to be implemented before it can be useful.

The scripts in [scripts/original_from_ydb_doc](scripts/original_from_ydb_doc) directory are copied from [YottaDB documentation](https://gitlab.com/YottaDB/DB/YDBDoc/tree/master/AdminOpsGuide/repl_procedures). Shout out to the great work by [YottaDB team](https://yottadb.com/)

### System Requirements

Tested with:

1. [Ubuntu 18.04](http://releases.ubuntu.com/18.04/)
2. [Go 1.13](https://golang.org/dl/)
3. [Confluent Platform 5.3](https://www.confluent.io/download/)
4. [MongoDB server](https://www.mongodb.com/download-center/community)
5. [MongoDB Kafka connector](https://www.confluent.io/hub/mongodb/kafka-connect-mongodb)
6. [YottaDB r1.28](https://yottadb.com/product/get-started/)
7. [FIS GT.M](https://en.wikipedia.org/wiki/GT.M)

### Build

```bash
go build ./cmd/cdcfilter

```
Build for AIX

```bash
GOOS=aix GOARCH=ppc64 go build ./cmd/cdcfilter
```

### Usage Guide 

#### ENV variables 
`GTMCDC_KAFKA_BROKERS`: kafka brokers list

`GTMCDC_KAFKA_TOPIC`: kafka topic

`GTMCDC_PROM_HTTP_ADDR`

`GTMCDC_LOG`: CDC process log file directory, defaults to `cdcfilter.log`

`GTMCDC_LOG_LEVEL`:  CDC process log level, defaults to `debug`

available logging level: 
- `panic`
- `fatal`
- `error`
- `warn`
- `warning`
- `info`
- `debug`
- `trace`

#### Using CDC Filter 
move compiled cdcfilter cmd to PATH 
```sh
mv ./cdcfilter /usr/bin/
```
allow execution for all users
```sh
chmod 777 /usr/bin/cdcfilter
```
start replication receiver process with filter
```sh
mupip replicate -receive -start -listenport="${GTM_RECEIVER_PORT}" -buffsize="${GTM_RECEIVER_BUFFER_SIZE} -log="${GTM_RECEIVER_LOG_PATH}" -filter=cdcfilter
```
