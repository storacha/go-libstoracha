package queue

import (
	"context"
	"iter"

	"github.com/ipni/go-libipni/ingest/schema"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/ipnipublisher/publisher"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-libstoracha/queuepoller"
)

type (
	PublishingJob struct {
		ProviderInfo peer.AddrInfo
		ContextID    string
		Digests      iter.Seq[mh.Multihash]
		Meta         metadata.Metadata
	}

	PublishingJobHandler struct {
		publisher publisher.AsyncPublisher
	}

	AdvertisementPublishingJobHandler struct {
		advertisementPublisher *publisher.AdvertisementPublisher
	}

	PublishingQueue queuepoller.Queue[PublishingJob]

	AdvertisementPublishingQueue queuepoller.Queue[schema.Advertisement]

	QueuePublisher struct {
		queue PublishingQueue
	}

	AdvertisementQueuePublisher struct {
		queue AdvertisementPublishingQueue
		store store.PublisherStore
	}
)

func NewPublishingJobHandler(publisher publisher.AsyncPublisher) *PublishingJobHandler {
	return &PublishingJobHandler{
		publisher: publisher,
	}
}

func (h *PublishingJobHandler) Handle(ctx context.Context, job PublishingJob) error {
	return h.publisher.Publish(ctx, job.ProviderInfo, job.ContextID, job.Digests, job.Meta)
}

func NewAdvertisementPublishingJobHandler(advertisementPublisher *publisher.AdvertisementPublisher) *AdvertisementPublishingJobHandler {
	return &AdvertisementPublishingJobHandler{
		advertisementPublisher: advertisementPublisher,
	}
}

func (h *AdvertisementPublishingJobHandler) Handle(ctx context.Context, jobs []queuepoller.WithID[schema.Advertisement]) map[string]error {
	for _, job := range jobs {
		h.advertisementPublisher.AddToBatch(job.Job)
	}
	_, err := h.advertisementPublisher.Commit(ctx)
	if err != nil {
		errs := make(map[string]error, len(jobs))
		for _, job := range jobs {
			errs[job.ID] = err
		}
		return errs
	}
	return map[string]error{}
}

var _ publisher.AsyncPublisher = (*QueuePublisher)(nil)

func NewQueuePublisher(queue PublishingQueue) *QueuePublisher {
	return &QueuePublisher{
		queue: queue,
	}
}

func (qp *QueuePublisher) Publish(ctx context.Context, pInfo peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) error {
	job := PublishingJob{
		ProviderInfo: pInfo,
		ContextID:    contextID,
		Digests:      digests,
		Meta:         meta,
	}
	return qp.queue.Queue(ctx, job)
}

type PublishingQueuePoller = queuepoller.QueuePoller[PublishingJob]

func NewPublishingQueuePoller(queue PublishingQueue, publisher publisher.AsyncPublisher, opts ...queuepoller.Option) (*PublishingQueuePoller, error) {
	handler := NewPublishingJobHandler(publisher)
	return queuepoller.NewQueuePoller(
		queue,
		queuepoller.BatchJobHandler(func(ctx context.Context, jobs []queuepoller.WithID[PublishingJob]) map[string]error {
			errs := make(map[string]error, len(jobs))
			for _, job := range jobs {
				err := handler.Handle(ctx, job.Job)
				if err != nil {
					errs[job.ID] = err
				}
			}
			return errs
		}),
		opts...)
}

var _ publisher.AsyncPublisher = (*AdvertisementQueuePublisher)(nil)

func NewAdvertisementQueuePublisher(queue AdvertisementPublishingQueue, store store.PublisherStore) *AdvertisementQueuePublisher {
	return &AdvertisementQueuePublisher{
		queue: queue,
		store: store,
	}
}

func (qa *AdvertisementQueuePublisher) Publish(ctx context.Context, pInfo peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) error {
	adv, err := publisher.GenerateAd(ctx, qa.store, pInfo.ID, pInfo.Addrs, []byte(contextID), meta, false, digests)
	if err != nil {
		return err
	}
	return qa.queue.Queue(ctx, adv)
}

type AdvertisementPublishingQueuePoller = queuepoller.QueuePoller[schema.Advertisement]

func NewAdvertisementPublishingQueuePoller(queue AdvertisementPublishingQueue, advertisementPublisher *publisher.AdvertisementPublisher, opts ...queuepoller.Option) (*AdvertisementPublishingQueuePoller, error) {
	return queuepoller.NewQueuePoller(
		queue,
		queuepoller.BatchJobHandler(NewAdvertisementPublishingJobHandler(advertisementPublisher).Handle),
		opts...)
}
