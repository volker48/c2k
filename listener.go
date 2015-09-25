package main

import (
	"github.com/volker48/c2k/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/kinesis"
	"io"
	"time"
)

type Listener struct {
	opts Options
	svc  *kinesis.Kinesis
}

func NewListener(opts Options, svc *kinesis.Kinesis) *Listener {
	return &Listener{opts: opts, svc: svc}
}

func (l *Listener) Listen(shardID, itrType, startingSequenceNumber string, wrtr io.Writer) {
	input := &kinesis.GetShardIteratorInput{ShardIteratorType: &itrType, ShardId: &shardID, StreamName: &l.opts.StreamName}
	if startingSequenceNumber != "" {
		input.StartingSequenceNumber = &startingSequenceNumber
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
		wrtr.Write(record.Data)
	}
	return recordsOut
}
