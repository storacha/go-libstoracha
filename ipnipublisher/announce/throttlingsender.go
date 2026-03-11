package announce

import (
	"context"
	"time"

	"github.com/ipni/go-libipni/announce"
	"github.com/ipni/go-libipni/announce/message"
	"github.com/storacha/go-libstoracha/throttler"
)

// ThrottlingSender is an [announce.Sender] that throttles the rate of sending
// announcements. It is intended to be used as a wrapper around another
// [announce.Sender] that will throttle the rate of sending announcements to the
// underlying sender.
type ThrottlingSender struct {
	sender announce.Sender
	action *throttler.Action[announceRequest]
}

type announceRequest struct {
	ctx context.Context
	msg message.Message
}

func NewThrottlingSender(sender announce.Sender, delay time.Duration) *ThrottlingSender {
	sendAnnounce := func(req announceRequest) error {
		return sender.Send(req.ctx, req.msg)
	}
	return &ThrottlingSender{
		sender: sender,
		action: throttler.NewAction(sendAnnounce, delay),
	}
}

func (s *ThrottlingSender) Close() error {
	s.action.Close()
	return s.sender.Close()
}

func (s *ThrottlingSender) Send(ctx context.Context, msg message.Message) error {
	return s.action.Execute(announceRequest{ctx: ctx, msg: msg})
}

var _ announce.Sender = (*ThrottlingSender)(nil)
