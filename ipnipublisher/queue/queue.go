package queue

import (
	"context"
	"iter"

	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/ipnipublisher/publisher"
)

type (
	PublishingJob struct {
		ID           string
		ProviderInfo peer.AddrInfo
		ContextID    string
		Digests      iter.Seq[mh.Multihash]
		Meta         metadata.Metadata
	}

	JobHandler struct {
		publisher publisher.Publisher
	}

	PublisherQueue interface {
		Queue(ctx context.Context, job PublishingJob) error
	}

	QueuePublisher struct {
		queue PublisherQueue
	}
)

func NewJobHandler(publisher publisher.Publisher) *JobHandler {
	return &JobHandler{
		publisher: publisher,
	}
}

func (j *JobHandler) Handle(ctx context.Context, job PublishingJob) error {
	_, err := j.publisher.Publish(ctx, job.ProviderInfo, job.ContextID, job.Digests, job.Meta)
	return err
}

var _ publisher.AsyncPublisher = (*QueuePublisher)(nil)

func NewQueuePublisher(queue PublisherQueue) *QueuePublisher {
	return &QueuePublisher{
		queue: queue,
	}
}

func (qp *QueuePublisher) Publish(ctx context.Context, pInfo peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) error {
	job := PublishingJob{
		ID:           "", // ID can be set by the queue implementation if needed.
		ProviderInfo: pInfo,
		ContextID:    contextID,
		Digests:      digests,
		Meta:         meta,
	}
	return qp.queue.Queue(ctx, job)
}
