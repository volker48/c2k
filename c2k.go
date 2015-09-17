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
	records := make([]*kinesis.PutRecordsRequestEntry, 500)
	rdr := bufio.NewReader(handle)
	i := 0
	for {
		line, err := rdr.ReadBytes([]byte(delimiter)[0])
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf(incompleteRead, fileName)
			break
		}
		records[i] = &kinesis.PutRecordsRequestEntry{Data: line, PartitionKey: &partitionKey}
		if i == 499 {
			shipAndCheck(svc, streamName, records)
			i = 0
		} else {
			i++
		}
	}
	if i != 0 {
		shipAndCheck(svc, streamName, records[:i])
	}
}

func shipAndCheck(svc *kinesis.Kinesis, streamName string, records []*kinesis.PutRecordsRequestEntry){
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
