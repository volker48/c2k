package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/satori/go.uuid"
	"log"
)

const MaxPutIdx = 499

type uploader struct {
	records  []*kinesis.PutRecordsRequestEntry
	position int
	opts     Options
	svc      *kinesis.Kinesis
}

type firehoseUploader struct {
	records  []*firehose.Record
	svc      *firehose.Firehose
	opts     Options
	position int
}

type Uploader interface {
	Upload(date []byte)
	Flush()
	shipAndCheck()
}

func NewUploader(svc *kinesis.Kinesis, fsvc *firehose.Firehose, opts Options) Uploader {
	if opts.Firehose {
		return &firehoseUploader{
			records:  make([]*firehose.Record, 500),
			position: 0,
			opts:     opts,
			svc:      fsvc,
		}
	} else {
		return &uploader{
			records:  make([]*kinesis.PutRecordsRequestEntry, 500),
			position: 0,
			opts:     opts,
			svc:      svc,
		}
	}

}

const MAX_PAYLOAD = 1000000

func pack(src []byte, dest *[]byte) bool {
	if len(src)+len(*dest)+1 > MAX_PAYLOAD {
		return true
	} else {
		if len(*dest) > 0 {
			*dest = append(*dest, '\n')
		}
		*dest = append(*dest, src...)
		return false
	}

}

func createRecord() *kinesis.PutRecordsRequestEntry {
	uid := uuid.NewV4().String()
	return &kinesis.PutRecordsRequestEntry{PartitionKey: &uid}
}

func (upldr *uploader) Upload(data []byte) {
	if upldr.records[upldr.position] == nil {
		log.Printf("Records at %d is nil creating new record", upldr.position)
		upldr.records[upldr.position] = createRecord()
	}
	full := pack(data, &upldr.records[upldr.position].Data)
	if full && upldr.position == MaxPutIdx {
		upldr.shipAndCheck()
		upldr.position = 0
		for i := 0; i < len(upldr.records); i++ {
			upldr.records[i] = nil
		}
	} else if full {
		upldr.position++
		record := createRecord()
		upldr.records[upldr.position] = record
		pack(data, &record.Data)
	}
}

func (fupldr *firehoseUploader) Upload(data []byte) {
	if fupldr.records[fupldr.position] == nil {
		fupldr.records[fupldr.position] = &firehose.Record{Data: data}
	}
	full := pack(data, &fupldr.records[fupldr.position].Data)
	if full && fupldr.position == MaxPutIdx {
		fupldr.shipAndCheck()
		fupldr.position = 0
		for i := 0; i < len(fupldr.records); i++ {
			fupldr.records[i] = nil
		}
	} else {
		fupldr.position++
	}
}

func (fupldr *firehoseUploader) shipAndCheck() {
	params := &firehose.PutRecordBatchInput{DeliveryStreamName: aws.String(fupldr.opts.StreamName)}
	if fupldr.position == MaxPutIdx {
		params.Records = fupldr.records
	} else {
		params.Records = fupldr.records[:fupldr.position+1]
	}
	resp, err := fupldr.svc.PutRecordBatch(params)
	if err != nil {
		log.Fatal("Error during firehose batch put ", err)
	}
	if *resp.FailedPutCount > 0 {
		log.Printf("%d records failed to upload", resp.FailedPutCount)
	}
	log.Printf("Successfully put %d record", int64(len(resp.RequestResponses))-*resp.FailedPutCount)
}

func (fupldr *firehoseUploader) Flush() {
	if fupldr.position == 0 && fupldr.records[fupldr.position] == nil {
		return
	}
	fupldr.shipAndCheck()
}

func (upldr *uploader) Flush() {
	if upldr.position == 0 && upldr.records[upldr.position] == nil {
		return
	}
	upldr.shipAndCheck()
}

func (upldr *uploader) shipAndCheck() {
	var records []*kinesis.PutRecordsRequestEntry
	if upldr.position == MaxPutIdx {
		records = upldr.records
	} else {
		records = upldr.records[:upldr.position+1]
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
