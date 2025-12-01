package mailgun

import "time"

// Config holds Mailgun transport settings.
type Config struct {
	ID           string        `yaml:"id" json:"id"`
	Domain       string        `yaml:"domain" json:"domain"`
	APIKey       string        `yaml:"api_key" json:"api_key"`
	SigningKey   string        `yaml:"signing_key" json:"signing_key"`
	BaseURL      string        `yaml:"base_url" json:"base_url"`
	RoutePrefix  string        `yaml:"route_prefix" json:"route_prefix"`
	AllowSenders []string      `yaml:"allow_senders" json:"allow_senders"`
	MaxBytes     int           `yaml:"max_bytes" json:"max_bytes"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`
}

// Defaults fills missing optional fields.
func (c *Config) Defaults() {
	if c.BaseURL == "" {
		c.BaseURL = "https://api.mailgun.net/v3"
	}
	if c.RoutePrefix == "" {
		c.RoutePrefix = ""
	}
	if c.MaxBytes == 0 {
		c.MaxBytes = 262144 // 256 KiB
	}
	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}
}

func (c *Config) Validate() error {
	if c.ID == "" {
		c.ID = "email-mailgun"
	}
	if c.Domain == "" {
		return Err("domain is required")
	}
	if c.APIKey == "" {
		return Err("api_key is required")
	}
	if c.SigningKey == "" {
		return Err("signing_key is required")
	}
	if len(c.AllowSenders) == 0 {
		return Err("allow_senders must include at least one sender")
	}
	return nil
}
