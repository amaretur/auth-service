package app

import (
	"os"
	"context"
	"os/signal"
)

func OnSignal(
	ctx context.Context,
	f func(),
	signals ...os.Signal,
) {

	nCtx, cancel := signal.NotifyContext(ctx, signals...)
	defer cancel()

	<-nCtx.Done()

	f()
}
