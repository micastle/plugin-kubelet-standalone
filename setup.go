package example

import (
	"go.uber.org/zap"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("example", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "example" and give us the next token.
	if c.NextArg() {
		// If there was another token, return an error, because we don't have any configuration.
		// Any errors returned from this setup function should be wrapped with plugin.Error, so we
		// can present a slightly nicer error message to the user.
		return plugin.Error("example", c.ArgErr())
	}

	InitStdOutLogger(zap.DebugLevel)

	logger, _ := GetLogger("Example")
	logger.Info("New Example created")
	config := GetDefaultConfig()
	tlsBypassHttpsClient := NewDefaultTlsBypassHttpsClient()
	client := NewClient(&config.Kubelet, tlsBypassHttpsClient)
	e := &Example{KubeClient: client, Logger: logger, Records: make([]*PodRecord, 0, 10)}

	go e.BackgroundLoop()

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		e.Next = next
		return e
	})

	// All OK, return a nil error.
	return nil
}
