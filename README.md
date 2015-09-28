# c2k - cat 2 kinesis
Like `netcat`, but for [Amazon Kinesis](https://aws.amazon.com/kinesis/).

## Usage

```
$>./c2k --help
Usage of ./c2k:
  -d string
    	Delimiter to split on (defaults to newline) (short) (default "\n")
  -delimiter string
    	Delimiter to split on (defaults to newline) (default "\n")
  -i string
    	Type of Shard Iterator to use. Valid choices: AT_SEQUENCE_NUMBER, AFTER_SEQUENCE_NUMBER, TRIM_HORIZON (short) (default "TRIM_HORIZON")
  -iter string
    	Type of Shard Iterator to use. Valid choices: AT_SEQUENCE_NUMBER, AFTER_SEQUENCE_NUMBER, TRIM_HORIZON (default "TRIM_HORIZON")
  -l	Listen to stream instead of sending data (short)
  -listen
    	Listen to stream instead of sending data
  -p string
    	AWS Profile name to use for authentication (short) (default "default")
  -partitionKey string
    	Partition key (default "1")
  -pk string
    	Partition key (short) (default "1")
  -profile string
    	AWS Profile name to use for authentication (default "default")
  -r string
    	AWS region, defaults to us-east-1 (short) (default "us-east-1")
  -region string
    	AWS region, defaults to us-east-1 (default "us-east-1")
  -s string
    	Stream name to put data (short)
  -sId string
    	Shard ID for listen purposes (short) (default "ALL")
  -shardId string
    	Shard ID for listen purposes (default "ALL")
  -sn string
    	Sequence number to use for iterators that use a sequence number (short)
  -startingSeqNum string
    	Sequence number to use for iterators that use a sequence number
  -streamName string
    	Stream name to put data
```


c2k operates in two modes. The first mode sends data to Kinesis. The second mode listens or reads data from a Kinesis stream.

### Sending data to kinesis
A common use case for c2k is sending each line of a file into Kinesis. Here is how you can do this with c2k:

```
c2k -s your-stream access.log error.log
```

This will put each line of `access.log` and `error.log` as a record into the Kinesis stream named `your-stream`. By default c2k will use the default crendentials in `~/.aws/credentials`, but you can choose the profile with the `-p` option.

### Listening for data
You can also listen for data in a Kinesis stream. By default c2k will listen to all shards in the stream, but you can specify a single shard id as well.

```
c2k -s your-stream -shardId 1
```

This will stream from shard id 1 of the stream named `your-stream`. c2k will write the data from the stream to standard out. By default, c2k will used the `TRIM_HORIZON` type of shard iterator.
