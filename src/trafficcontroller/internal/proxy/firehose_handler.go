package proxy

import (
	"context"
	"log"
	"net/http"
	"plumbing"
	"sync/atomic"

	"github.com/gorilla/mux"
)

const firehoseID = "firehose"

type FirehoseHandler struct {
	grpcConn grpcConnector
	counter  int64
}

func NewFirehoseHandler(grpcConn grpcConnector) *FirehoseHandler {
	return &FirehoseHandler{grpcConn: grpcConn}
}

func (h *FirehoseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&h.counter, 1)
	defer atomic.AddInt64(&h.counter, -1)

	subID := mux.Vars(r)["subID"]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := h.grpcConn.Subscribe(ctx, &plumbing.SubscriptionRequest{
		ShardID: subID,
	})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Printf("error occurred when subscribing to doppler: %s", err)
		return
	}

	serveWS(firehoseID, subID, w, r, client)
}

func (h *FirehoseHandler) Count() int64 {
	return atomic.LoadInt64(&h.counter)
}