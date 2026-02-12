package sentryClient

import (
	"log"
	"os"
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
		os.Exit(3)
	}
	defer sentry.Flush(2 * time.Second)

	log.Println("sentry loaded successfully")
}

func NewSentryClient(sentryConfig sentry.ClientOptions) SentryClient {
	return &sentryClient{
		config: sentryConfig,
	}
}
