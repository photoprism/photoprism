package commands

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

// StatusCommand configures the command name, flags, and action.
var StatusCommand = &cli.Command{
	Name:   "status",
	Usage:  "Checks if the Web server is running",
	Action: statusAction,
}

// statusAction checks if the Web server is running.
func statusAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	// Create new http.Client instance.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	client := &http.Client{Timeout: 10 * time.Second}

	// Connect to unix socket?
	if unixSocket := conf.HttpSocket(); unixSocket != nil {
		client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial(unixSocket.Scheme, unixSocket.Path)
			},
		}
	}

	endpointUrl := buildStatusEndpoint(conf)

	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)

	if err != nil {
		return err
	}

	var response string

	if resp, reqErr := client.Do(req); reqErr != nil {
		return fmt.Errorf("cannot connect to %s:%d", conf.HttpHost(), conf.HttpPort())
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("server running at %s:%d, bad status %d\n", conf.HttpHost(), conf.HttpPort(), resp.StatusCode)
	} else if body, readErr := io.ReadAll(resp.Body); readErr != nil {
		return readErr
	} else {
		response = string(body)
	}

	message := gjson.Get(response, "status").String()

	if message != "" {
		fmt.Println(message)
	} else {
		fmt.Println("unknown")
	}

	return nil
}

// buildStatusEndpoint returns the status endpoint URL, preferring the public
// SiteUrl (which carries the correct scheme) and falling back to the local
// HTTP host/port. When a Unix socket is configured, an http+unix style URL is
// used so the custom transport can dial the socket.
func buildStatusEndpoint(conf *config.Config) string {
	if socket := conf.HttpSocket(); socket != nil {
		return fmt.Sprintf("%s://%s/api/v1/status", socket.Scheme, strings.TrimPrefix(socket.Path, "/"))
	}

	siteUrl := strings.TrimRight(conf.SiteUrl(), "/")

	if siteUrl != "" {
		return siteUrl + "/api/v1/status"
	}

	return fmt.Sprintf("http://%s:%d/api/v1/status", conf.HttpHost(), conf.HttpPort())
}
