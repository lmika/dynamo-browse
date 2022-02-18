package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/lmika/gopkgs/cli"
)

func main() {
	flagQueue := flag.String("q", "", "URL of queue to drain")
	flagDir := flag.String("dir", "", "directory to save messages")
	flag.Parse()

	if *flagQueue == "" {
		cli.Fatalf("-q flag needs to be specified")
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	outDir := *flagDir
	if outDir == "" {
		outDir = "out-" + time.Now().Format("20060102150405")
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		cli.Fatalf("unable to create out dir: %v", err)
	}

	client := sqs.NewFromConfig(cfg)
	msgCount := 0
	for {
		out, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(*flagQueue),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds: 1,
		})
		if err != nil {
			log.Fatalf("error receiving messages: %v", err)
			break
		} else if len(out.Messages) == 0 {
			break
		}

		messagesToDelete := make([]types.DeleteMessageBatchRequestEntry, 0, 10)
		for _, msg := range out.Messages {
			if err := handleMessage(ctx, outDir, msg); err == nil {
				messagesToDelete = append(messagesToDelete, types.DeleteMessageBatchRequestEntry{
					Id: msg.MessageId,
					ReceiptHandle: msg.ReceiptHandle,
				})
				msgCount += 1
			} else {
				log.Println(err)
			}
		}
		if len(messagesToDelete) == 0 {
			log.Printf("no messages handled, terminating")
			break
		}

		if _, err := client.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
			QueueUrl: aws.String(*flagQueue),
			Entries: messagesToDelete,
		}); err != nil {
			log.Printf("error deleting messages from queue: %v", err)
			break
		}
	}

	log.Printf("Handled %v messages", msgCount)
}

func handleMessage(ctx context.Context, outDir string, msg types.Message) error {
	outFile := filepath.Join(outDir, aws.ToString(msg.MessageId) + ".json")
	msgBody := aws.ToString(msg.Body)

	log.Printf("%v -> %v", aws.ToString(msg.MessageId), outFile)
	if err := os.WriteFile(outFile, []byte(msgBody), 0644); err != nil {
		return errors.Wrapf(err, "unable to write message %v to file %v", msg.MessageId, outFile)
	}

	return nil
}
