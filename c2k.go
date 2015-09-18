package main
import (
	"os"
	"bufio"
	"flag"
	"io"
	"log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/aws/credentials"
)
const (
	defaultDelimiter = "\n"
	delimiterUsage = "Delimiter to split on (defaults to newline)"
	defaultProfile = "default"
	profileUsage = "AWS Profile name to use for authentication"
	defaultPartitionKey = "1"
	partitionKeyUsage = "Partition key"
	defaultRegion = "us-east-1"
	regionUsage = "AWS region, defaults to us-east-1"
	streamNameUsage = "Stream name to put data"
	incompleteRead = "c2k: incomplete read of stream"
	noSuchFile = "c2k: %s: no such file"
)

type Options struct {
	Delimiter, Profile, Region, StreamName, PartitionKey string
}

func main() {
	opts := parseArgs()

	svc := createService(opts.Profile, opts.Region)

	for _, fileName := range (flag.Args()) {
		uploadFile(fileName, opts, svc)
	}
}

func createService(profile, region string) *kinesis.Kinesis {
	creds := credentials.NewSharedCredentials("", profile)
	return kinesis.New(&aws.Config{Region: aws.String(region), Credentials: creds})
}

func parseArgs() Options {
	opts := Options{}
	flag.StringVar(&opts.Delimiter, "delimiter", defaultDelimiter, delimiterUsage)
	flag.StringVar(&opts.Delimiter, "d", defaultDelimiter, delimiterUsage + " (short)")
	flag.StringVar(&opts.PartitionKey, "partitionKey", defaultPartitionKey, partitionKeyUsage)
	flag.StringVar(&opts.PartitionKey, "pk", defaultPartitionKey, partitionKeyUsage + " (short)")
	flag.StringVar(&opts.Profile, "profile", defaultProfile, profileUsage)
	flag.StringVar(&opts.Profile, "p", defaultProfile, profileUsage + " (short)")
	flag.StringVar(&opts.Region, "region", defaultRegion, regionUsage)
	flag.StringVar(&opts.Region, "r", defaultRegion, regionUsage + " (short)")
	flag.StringVar(&opts.StreamName, "streamName", "", streamNameUsage)
	flag.StringVar(&opts.StreamName, "s", "", streamNameUsage + " (short)")
	flag.Parse()
	if opts.StreamName == "" {
		log.Fatal("streamName is a required parameter")
	}
	return opts
}

func uploadFile(fileName string, opts Options, svc *kinesis.Kinesis) {
	handle, err := os.Open(fileName)
	if err != nil {
		log.Printf(noSuchFile, fileName)
		return
	}
	defer handle.Close()
	rdr := bufio.NewReader(handle)
	putFromReader(rdr, opts, svc)
}

func putFromReader(rdr *bufio.Reader, opts Options, svc *kinesis.Kinesis) {
	records := make([]*kinesis.PutRecordsRequestEntry, 500)
	i := 0
	for {
		line, err := rdr.ReadBytes([]byte(opts.Delimiter)[0])
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf(incompleteRead)
			break
		}
		records[i] = &kinesis.PutRecordsRequestEntry{Data: line, PartitionKey: &opts.PartitionKey}
		if i == 499 {
			shipAndCheck(svc, opts.StreamName, records)
			i = 0
		} else {
			i++
		}
	}
	if i != 0 {
		shipAndCheck(svc, opts.StreamName, records[:i])
	}
}

func shipAndCheck(svc *kinesis.Kinesis, streamName string, records []*kinesis.PutRecordsRequestEntry) {
	putRecordsOutput, err := shipRecords(svc, streamName, records)
	if err != nil {
		log.Fatal("Error during put records ", err)
	}

	if *putRecordsOutput.FailedRecordCount > 0 {
		//TODO: Decide what to do with failures. Either write failures to a new failures file or retry?
		log.Printf("%d records failed to upload", putRecordsOutput.FailedRecordCount)
	}

	log.Printf("Successfully put %d records", int64(len(putRecordsOutput.Records)) - *putRecordsOutput.FailedRecordCount)
}

func shipRecords(svc *kinesis.Kinesis, streamName string, records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	putRecordsInput := &kinesis.PutRecordsInput{Records: records, StreamName: &streamName}
	return svc.PutRecords(putRecordsInput)
}
