package httpadapter

import "expvar"

var (
	webhookRetryEnqueued  = expvar.NewInt("webhook_retry_enqueued")
	webhookRetrySucceeded = expvar.NewInt("webhook_retry_succeeded")
	webhookRetryFailed    = expvar.NewInt("webhook_retry_failed")
	webhookRetryDead      = expvar.NewInt("webhook_retry_dead")
	webhookRetryLastRun   = expvar.NewString("webhook_retry_last_run")
)
