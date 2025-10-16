package awsutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

// queueMessage is the struct that is serialized onto an SQS message queue in JSON
type queueMessage[Message any] struct {
	JobID   uuid.UUID `json:"JobID,omitempty"`
	Message Message   `json:"Message,omitempty"`
}

// SerializedJob represents a job that has been serialized for transport over SQS + S3
type SerializedJob[Message any] struct {
	ID       string
	GroupID  *string
	Message  Message
	Extended io.Reader
}

type JobMarshaller[Job any, Message any] interface {
	Marshall(job Job) (SerializedJob[Message], error)
	Unmarshall(SerializedJob[Message]) (Job, error)
	Empty() Job
}

// SQSExtendedQueue implements a queue interface using SQS that can store extended data to an S3 bucket
type SQSExtendedQueue[Job any, Message any] struct {
	queueID    string
	bucket     string
	s3Client   *s3.Client
	sqsClient  *sqs.Client
	decoder    *SQSDecoder[Job, Message]
	marshaller JobMarshaller[Job, Message]
}

// NewSQSExtendedQueue returns a new SQSExtendedQueue for the given aws config
func NewSQSExtendedQueue[Job any, Message any](cfg aws.Config, queueID string, bucket string, marshaller JobMarshaller[Job, Message]) *SQSExtendedQueue[Job, Message] {
	return &SQSExtendedQueue[Job, Message]{
		queueID:    queueID,
		bucket:     bucket,
		s3Client:   s3.NewFromConfig(cfg),
		sqsClient:  sqs.NewFromConfig(cfg),
		marshaller: marshaller,
		decoder:    NewSQSDecoder(cfg, bucket, marshaller),
	}
}

// Queue implements blobindexlookup.CachingQueue.
func (s *SQSExtendedQueue[Job, Message]) Queue(ctx context.Context, job Job) error {
	uuid := uuid.New()
	jobMessage, err := s.marshaller.Marshall(job)
	if err != nil {
		return fmt.Errorf("marshalling job: %w", err)
	}
	data, err := io.ReadAll(jobMessage.Extended)
	if err != nil {
		return fmt.Errorf("reading message: %w", err)
	}
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(uuid.String()),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	})
	if err != nil {
		return fmt.Errorf("saving index CAR to S3: %w", err)
	}
	err = s.sendMessage(ctx, jobMessage.GroupID, queueMessage[Message]{
		JobID:   uuid,
		Message: jobMessage.Message,
	})
	if err != nil {
		// error sending message so cleanup queue
		_, s3deleteErr := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(uuid.String()),
		})
		if s3deleteErr != nil {
			err = errors.Join(err, fmt.Errorf("cleaning up index CAR on S3: %w", s3deleteErr))
		}
	}
	return err
}

func (s *SQSExtendedQueue[Job, Message]) sendMessage(ctx context.Context, groupID *string, msg queueMessage[Message]) error {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("serializing message json: %w", err)
	}
	message := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueID),
		MessageBody: aws.String(string(messageJSON)),
	}
	if groupID != nil {
		message.MessageGroupId = groupID
	}
	_, err = s.sqsClient.SendMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("enqueueing message: %w", err)
	}
	return nil
}

// Read reads a batch of jobs from the SQS queue.
// Returns an empty slice if no jobs are available.
// The caller must process jobs and delete them from the queue when done.
func (s *SQSExtendedQueue[Job, Message]) Read(ctx context.Context, maxJobs int) ([]Job, error) {
	receiveOutput, err := s.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueID),
		MaxNumberOfMessages: int32(maxJobs),
		WaitTimeSeconds:     20, // enable long-polling
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	if len(receiveOutput.Messages) == 0 {
		return []Job{}, nil
	}

	jobs := make([]Job, 0, len(receiveOutput.Messages))
	for _, msg := range receiveOutput.Messages {
		job, err := s.decoder.DecodeMessage(ctx, aws.ToString(msg.ReceiptHandle), aws.ToString(msg.Body))
		if err != nil {
			return nil, fmt.Errorf("failed to decode message: %w", err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Release makes a job available for processing again by making it visible in the queue
func (s *SQSExtendedQueue[Job, Message]) Release(ctx context.Context, jobID string) error {
	_, err := s.sqsClient.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.queueID),
		ReceiptHandle:     aws.String(jobID),
		VisibilityTimeout: 0,
	})

	return err
}

// Delete deletes a job message from the SQS queue.
func (s *SQSExtendedQueue[Job, Message]) Delete(ctx context.Context, jobID string) error {
	_, err := s.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueID),
		ReceiptHandle: aws.String(jobID),
	})
	return err
}

// SQSDecoder provides interfaces for working with jobs received over SQS
type SQSDecoder[Job any, Message any] struct {
	bucket     string
	s3Client   *s3.Client
	marshaller JobMarshaller[Job, Message]
}

// NewSQSDecoder returns a new decoder for the given AWS config
func NewSQSDecoder[Job any, Message any](cfg aws.Config, bucket string, marshaller JobMarshaller[Job, Message]) *SQSDecoder[Job, Message] {
	return &SQSDecoder[Job, Message]{
		bucket:     bucket,
		s3Client:   s3.NewFromConfig(cfg),
		marshaller: marshaller,
	}
}

// DecodeMessage decodes a provider caching job from the SQS message body, reading the stored index from S3
func (s *SQSDecoder[Job, Message]) DecodeMessage(ctx context.Context, receiptHandle string, messageBody string) (Job, error) {
	var msg queueMessage[Message]
	err := json.Unmarshal([]byte(messageBody), &msg)
	if err != nil {
		return s.marshaller.Empty(), fmt.Errorf("deserializing message: %w", err)
	}
	received, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(msg.JobID.String()),
	})
	if err != nil {
		return s.marshaller.Empty(), fmt.Errorf("reading stored index CAR: %w", err)
	}
	defer received.Body.Close()
	return s.marshaller.Unmarshall(SerializedJob[Message]{
		ID:       receiptHandle,
		Message:  msg.Message,
		Extended: received.Body,
	})
}
