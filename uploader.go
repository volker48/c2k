package main

import (
	"github.com/volker48/c2k/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/kinesis"
	"log"
)

const MaxPutIdx = 499

type Uploader struct {
	records  []*kinesis.PutRecordsRequestEntry
	position int
	opts     Options
	svc      *kinesis.Kinesis
}

func NewUploader(svc *kinesis.Kinesis, opts Options) *Uploader {
	return &Uploader{
		records:  make([]*kinesis.PutRecordsRequestEntry, 500),
		position: 0,
		opts:     opts,
		svc:      svc,
	}
}

func (upldr *Uploader) Upload(data []byte) {
	requestEntry := &kinesis.PutRecordsRequestEntry{Data: data, PartitionKey: &upldr.opts.PartitionKey}
	upldr.records[upldr.position] = requestEntry
	if upldr.position == MaxPutIdx {
		upldr.shipAndCheck()
		upldr.position = 0
	} else {
		upldr.position++
	}
}

func (upldr *Uploader) Flush() {
	if upldr.position == 0 {
		return
	}
	upldr.shipAndCheck()
}

func (upldr *Uploader) shipAndCheck() {
	var records []*kinesis.PutRecordsRequestEntry
	if upldr.position == MaxPutIdx {
		records = upldr.records
	} else {
		records = upldr.records[:upldr.position]
	}
	putRecordsInput := &kinesis.PutRecordsInput{Records: records, StreamName: &upldr.opts.StreamName}
	putRecordsOutput, err := upldr.svc.PutRecords(putRecordsInput)
	if err != nil {
		log.Fatal("Error during put records ", err)
	}

	if *putRecordsOutput.FailedRecordCount > 0 {
		//TODO: Decide what to do with failures. Either write failures to a new failures file or retry?
		log.Printf("%d records failed to upload", putRecordsOutput.FailedRecordCount)
	}

	log.Printf("Successfully put %d records", int64(len(putRecordsOutput.Records))-*putRecordsOutput.FailedRecordCount)
}
