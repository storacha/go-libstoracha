package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ipni/go-libipni/find/model"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/awsutils"
	"github.com/storacha/go-libstoracha/ipnipublisher/queue"
	"github.com/storacha/go-libstoracha/metadata"
)

// SQSPublishingQueue implements the a queue for publishing jobs on SQS
type SQSPublishingQueue = awsutils.SQSExtendedQueue[queue.PublishingJob, model.ProviderResult]

type jobMarshaller struct{}

func (jm jobMarshaller) Marshall(job queue.PublishingJob) (awsutils.SerializedJob[model.ProviderResult], error) {
	digests := slices.Collect(job.Digests)
	data, err := json.Marshal(digests)
	if err != nil {
		return awsutils.SerializedJob[model.ProviderResult]{}, fmt.Errorf("serializing digests to json: %w", err)
	}
	reader := bytes.NewReader(data)
	metaBytes, err := job.Meta.MarshalBinary()
	if err != nil {
		return awsutils.SerializedJob[model.ProviderResult]{}, fmt.Errorf("serializing metadata to binary: %w", err)
	}
	return awsutils.SerializedJob[model.ProviderResult]{
		ID: job.ID,
		Message: model.ProviderResult{
			Provider:  &job.ProviderInfo,
			ContextID: []byte(job.ContextID),
			Metadata:  metaBytes,
		},
		Extended: reader,
	}, nil
}

func (jm jobMarshaller) Unmarshall(sj awsutils.SerializedJob[model.ProviderResult]) (queue.PublishingJob, error) {
	digests := []mh.Multihash{}
	err := json.NewDecoder(sj.Extended).Decode(&digests)
	if err != nil {
		return queue.PublishingJob{}, fmt.Errorf("deserializing index from CAR: %w", err)
	}
	metadata := metadata.MetadataContext.New()
	err = metadata.UnmarshalBinary(sj.Message.Metadata)

	if err != nil {
		return queue.PublishingJob{}, fmt.Errorf("deserializing metadata from binary: %w", err)
	}

	return queue.PublishingJob{
		ID:           sj.ID,
		ProviderInfo: *sj.Message.Provider,
		ContextID:    string(sj.Message.ContextID),
		Meta:         metadata,
		Digests:      slices.Values(digests),
	}, nil
}

func (jm jobMarshaller) Empty() queue.PublishingJob {
	return queue.PublishingJob{}
}

// NewSQSPublishingQueue returns a new SQSPublishingQueue for the given aws config
func NewSQSPublishingQueue(cfg aws.Config, queueID string, bucket string) *SQSPublishingQueue {
	return awsutils.NewSQSExtendedQueue(cfg, queueID, bucket, jobMarshaller{})
}
