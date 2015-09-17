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
	incompleteRead = "c2k: %s: incomplete read"
	noSuchFile = "c2k: %s: no such file"
)

//var delimiter, profile, region, streamName, partitionKey string

func main() {
	delimiter, profile, region, streamName, partitionKey := parseArgs()

	svc := createService(profile, region)


	for _, fileName := range (flag.Args()) {
		uploadFile(fileName, delimiter, streamName, partitionKey, svc)
	}
}

func createService(profile, region string) *kinesis.Kinesis {
	creds := credentials.NewSharedCredentials("", profile)
	return kinesis.New(&aws.Config{Region: aws.String(region), Credentials: creds})
}

func parseArgs() (delimiter, profile, region, streamName, partitionKey string) {
	flag.StringVar(&delimiter, "delimiter", defaultDelimiter, delimiterUsage)
	flag.StringVar(&delimiter, "d", defaultDelimiter, delimiterUsage + " (short)")
	flag.StringVar(&partitionKey, "partitionKey", defaultPartitionKey, partitionKeyUsage)
	flag.StringVar(&partitionKey, "pk", defaultPartitionKey, partitionKeyUsage + " (short)")
	flag.StringVar(&profile, "profile", defaultProfile, profileUsage)
	flag.StringVar(&profile, "p", defaultProfile, profileUsage + " (short)")
	flag.StringVar(&region, "region", defaultRegion, regionUsage)
	flag.StringVar(&region, "r", defaultRegion, regionUsage + " (short)")
	flag.StringVar(&streamName, "streamName", "", streamNameUsage)
	flag.StringVar(&streamName, "s", "", streamNameUsage + " (short)")
	flag.Parse()
	if streamName == "" {
		log.Fatal("streamName is a required parameter")
	}
	return
}

func uploadFile(fileName, delimiter, streamName, partitionKey string, svc *kinesis.Kinesis) {
	handle, err := os.Open(fileName)
	if err != nil {
		log.Printf(noSuchFile, fileName)
		return
	}
	defer handle.Close()
	records := make([]*kinesis.PutRecordsRequestEntry, 0)
	rdr := bufio.NewReader(handle)
	for {
		line, err := rdr.ReadBytes([]byte(delimiter)[0])
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf(incompleteRead, fileName)
			break
		}
		records = append(records, &kinesis.PutRecordsRequestEntry{Data: line, PartitionKey: &partitionKey})
	}
	putRecordsInput := &kinesis.PutRecordsInput{Records: records, StreamName: &streamName}
	putRecordsOutput, err := svc.PutRecords(putRecordsInput)
	if err != nil {
		log.Fatal("Error during put records ", err)
	}

	if *putRecordsOutput.FailedRecordCount > 0 {
		log.Printf("%d records failed to upload", putRecordsOutput.FailedRecordCount)
	}

	log.Printf("Successfully put %d records", int64(len(putRecordsOutput.Records)) - *putRecordsOutput.FailedRecordCount)
}
