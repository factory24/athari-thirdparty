package sentryClient

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

type SentryClient interface {
	Connect()
}

type sentryClient struct {
	config sentry.ClientOptions
}

func (cfg *sentryClient) Connect() {
	log.Println("setting up sentry")

	if err := sentry.Init(cfg.config); err != nil {
		log.Println("sentry.Init", err)
	} else {
		log.Println("sentry loaded successfully")
	}
}

func NewSentryClient(sentryConfig sentry.ClientOptions) SentryClient {
	return &sentryClient{
		config: sentryConfig,
	}
}

func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

func CaptureException(err error) {
	sentry.CaptureException(err)
}

func CaptureMessage(msg string) {
	sentry.CaptureMessage(msg)
}
