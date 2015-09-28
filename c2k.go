package main

import (
	"bufio"
	"flag"
	"github.com/volker48/c2k/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/volker48/c2k/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/volker48/c2k/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/kinesis"
	"io"
	"log"
	"os"
)

const (
	defaultDelimiter           = "\n"
	delimiterUsage             = "Delimiter to split on (defaults to newline)"
	ItrUsage                   = "Type of Shard Iterator to use. Valid choices: AT_SEQUENCE_NUMBER, AFTER_SEQUENCE_NUMBER, TRIM_HORIZON"
	listenUsage                = "Listen to stream instead of sending data"
	defaultProfile             = "default"
	profileUsage               = "AWS Profile name to use for authentication"
	defaultPartitionKey        = "1"
	partitionKeyUsage          = "Partition key"
	defaultRegion              = "us-east-1"
	regionUsage                = "AWS region, defaults to us-east-1"
	startingSeqNumUsage        = "Sequence number to use for iterators that use a sequence number"
	shardIdUsage               = "Shard ID for listen purposes"
	defaultShardId      string = "ALL"
	streamNameUsage            = "Stream name to put data"
	incompleteRead             = "c2k: incomplete read of stream"
	noSuchFile                 = "c2k: %s: no such file"
	TrimHorizon         string = "TRIM_HORIZON"
	AtSequenceNum       string = "AT_SEQUENCE_NUMBER"
	AfterSequenceNum    string = "AFTER_SEQUENCE_NUMBER"
	Latest              string = "LATEST"
)

type Options struct {
	Delimiter, Profile, Region, ShardId, StartingSeqNum, StreamName, PartitionKey, ItrType string
}

func main() {
	var listen bool
	opts := parseArgs(&listen)

	svc := createService(opts.Profile, opts.Region)

	if listen {
		listener := NewListener(opts, svc)
		listener.Listen(os.Stdout)
	} else {
		for _, fileName := range flag.Args() {
			uploadFile(fileName, opts, svc)
		}
	}

}

func createService(profile, region string) *kinesis.Kinesis {
	creds := credentials.NewSharedCredentials("", profile)
	return kinesis.New(&aws.Config{Region: aws.String(region), Credentials: creds})
}

func parseArgs(listen *bool) Options {
	opts := Options{}
	flag.StringVar(&opts.Delimiter, "delimiter", defaultDelimiter, delimiterUsage)
	flag.StringVar(&opts.Delimiter, "d", defaultDelimiter, delimiterUsage+" (short)")
	flag.StringVar(&opts.ItrType, "iter", TrimHorizon, ItrUsage)
	flag.StringVar(&opts.ItrType, "i", TrimHorizon, ItrUsage+" (short)")
	flag.BoolVar(listen, "listen", false, listenUsage)
	flag.BoolVar(listen, "l", false, listenUsage+" (short)")
	flag.StringVar(&opts.PartitionKey, "partitionKey", defaultPartitionKey, partitionKeyUsage)
	flag.StringVar(&opts.PartitionKey, "pk", defaultPartitionKey, partitionKeyUsage+" (short)")
	flag.StringVar(&opts.Profile, "profile", defaultProfile, profileUsage)
	flag.StringVar(&opts.Profile, "p", defaultProfile, profileUsage+" (short)")
	flag.StringVar(&opts.Region, "region", defaultRegion, regionUsage)
	flag.StringVar(&opts.Region, "r", defaultRegion, regionUsage+" (short)")
	flag.StringVar(&opts.StartingSeqNum, "startingSeqNum", "", startingSeqNumUsage)
	flag.StringVar(&opts.StartingSeqNum, "sn", "", startingSeqNumUsage+" (short)")
	flag.StringVar(&opts.ShardId, "shardId", defaultShardId, shardIdUsage)
	flag.StringVar(&opts.ShardId, "sId", defaultShardId, shardIdUsage+" (short)")
	flag.StringVar(&opts.StreamName, "streamName", "", streamNameUsage)
	flag.StringVar(&opts.StreamName, "s", "", streamNameUsage+" (short)")
	flag.Parse()
	if opts.StreamName == "" {
		log.Fatal("streamName is a required parameter")
	}
	if *listen && opts.ItrType != TrimHorizon && opts.ItrType != Latest && opts.ItrType != AfterSequenceNum && opts.ItrType != AtSequenceNum {
		log.Fatal("Invalid iter type given ", opts.ItrType)
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
	uploader := NewUploader(svc, opts)
	for {
		data, err := rdr.ReadBytes([]byte(opts.Delimiter)[0])
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf(incompleteRead)
			break
		}
		uploader.Upload(data)
	}
	uploader.Flush()
}
