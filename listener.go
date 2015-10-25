package main

import (
	"bufio"
	"bytes"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

type Listener struct {
	opts Options
	svc  *kinesis.Kinesis
}

func NewListener(opts Options, svc *kinesis.Kinesis) *Listener {
	return &Listener{opts: opts, svc: svc}
}

func getShardIds(svc *kinesis.Kinesis, streamName string) []*kinesis.Shard {
	var shards []*kinesis.Shard
	describeInput := &kinesis.DescribeStreamInput{StreamName: &streamName}
	for {
		out, err := svc.DescribeStream(describeInput)
		if err != nil {
			log.Fatal("Could not describe stream to find shard ids: ", err)
		}
		shards = append(shards, out.StreamDescription.Shards...)
		if !*out.StreamDescription.HasMoreShards {
			break
		}
		lastShard := out.StreamDescription.Shards[len(out.StreamDescription.Shards)-1]
		describeInput.ExclusiveStartShardId = lastShard.ShardId
	}

	return shards
}

func (l *Listener) Listen(wrtr io.Writer) {
	var shards []*kinesis.Shard
	if l.opts.ShardId == defaultShardId {
		shards = getShardIds(l.svc, l.opts.StreamName)
	} else {
		shards = []*kinesis.Shard{&kinesis.Shard{ShardId: &l.opts.ShardId}}
	}
	for _, shard := range shards {
		go l.followIterator(*shard.ShardId, wrtr)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Printf("Received signal: %s", s)
}

func (l *Listener) followIterator(shardId string, wrtr io.Writer) {
	input := &kinesis.GetShardIteratorInput{ShardIteratorType: &l.opts.ItrType, ShardId: &shardId, StreamName: &l.opts.StreamName}
	if l.opts.StartingSeqNum != "" {
		input.StartingSequenceNumber = &l.opts.StartingSeqNum
	}
	out, err := l.svc.GetShardIterator(input)
	if err != nil {
		//TODO: Actually handle this
		panic(err)
	}
	shardIterator := out.ShardIterator
	for {
		recordsOut := l.writeRecords(shardIterator, wrtr)
		shardIterator = recordsOut.NextShardIterator
		// If we are caught up just wait
		if *recordsOut.MillisBehindLatest == 0 {
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *Listener) writeRecords(shardIterator *string, wrtr io.Writer) (recordsOut *kinesis.GetRecordsOutput) {
	getInput := &kinesis.GetRecordsInput{ShardIterator: shardIterator}
	recordsOut, err := l.svc.GetRecords(getInput)
	if err != nil {
		//TODO: Handle it
		panic(err)
	}
	for _, record := range recordsOut.Records {
		brdr := bufio.NewScanner(bytes.NewReader(record.Data))
		for brdr.Scan() {
			err := brdr.Err()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("Error reading record data: ", err)
			}
			//These lines are for removing blank lines from data payloads
			data := bytes.TrimSpace(brdr.Bytes())
			if len(data) == 0 {
				continue
			}
			data = append(data, '\n')
			wrtr.Write(data)
		}
	}
	return recordsOut
}
