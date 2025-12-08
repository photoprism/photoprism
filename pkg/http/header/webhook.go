package header

const (
	// WebhookID is the request header containing a webhook identifier.
	WebhookID string = "webhook-id"
	// WebhookSignature carries the signature header.
	WebhookSignature string = "webhook-signature"
	// WebhookTimestamp carries the timestamp header.
	WebhookTimestamp string = "webhook-timestamp"
	// WebhookSecretPrefix prefixes stored webhook secrets.
	WebhookSecretPrefix string = "whsec_"
)
