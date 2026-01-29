package confluence

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const baseURLv2 = "/wiki/api/v2"

// Client is a Confluence API client.
type Client struct {
	transport http.RoundTripper
	insecure  bool
	server    string
	login     string
	authType  *jira.AuthType
	token     string
	timeout   time.Duration
	debug     bool
}

// ClientFunc decorates option for client.
type ClientFunc func(*Client)

// NewClient instantiates a new Confluence client.
func NewClient(c jira.Config, opts ...ClientFunc) *Client {
	client := Client{
		server:   strings.TrimSuffix(c.Server, "/"),
		login:    c.Login,
		token:    c.APIToken,
		authType: c.AuthType,
		debug:    c.Debug,
	}

	for _, opt := range opts {
		opt(&client)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: client.insecure,
		},
		DialContext: (&net.Dialer{
			Timeout: client.timeout,
		}).DialContext,
	}

	if c.AuthType != nil && *c.AuthType == jira.AuthTypeMTLS {
		caCert, err := os.ReadFile(c.MTLSConfig.CaCert)
		if err != nil {
			log.Fatalf("%s, %s", err, c.MTLSConfig.CaCert)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(c.MTLSConfig.ClientCert, c.MTLSConfig.ClientKey)
		if err != nil {
			log.Fatal(err)
		}

		transport.TLSClientConfig.RootCAs = caCertPool
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
		transport.TLSClientConfig.Renegotiation = tls.RenegotiateFreelyAsClient
	}

	client.transport = transport

	return &client
}

// WithTimeout is a functional opt to attach timeout to the client.
func WithTimeout(to time.Duration) ClientFunc {
	return func(c *Client) {
		c.timeout = to
	}
}

// WithInsecureTLS is a functional opt that allows you to skip TLS certificate verification.
func WithInsecureTLS(ins bool) ClientFunc {
	return func(c *Client) {
		c.insecure = ins
	}
}

// Get sends a GET request to the Confluence v2 API.
func (c *Client) Get(ctx context.Context, path string, headers jira.Header) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, c.server+baseURLv2+path, nil, headers)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body []byte, headers jira.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var res *http.Response

	defer func() {
		if c.debug {
			dump(req, res)
		}
	}()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.authType == nil {
		basic := jira.AuthTypeBasic
		c.authType = &basic
	}

	switch c.authType.String() {
	case string(jira.AuthTypeMTLS):
		if c.token != "" {
			req.Header.Add("Authorization", "Bearer "+c.token)
		}
	case string(jira.AuthTypeBearer):
		req.Header.Add("Authorization", "Bearer "+c.token)
	case string(jira.AuthTypeBasic):
		req.SetBasicAuth(c.login, c.token)
	}

	httpClient := &http.Client{Transport: c.transport}

	return httpClient.Do(req.WithContext(ctx))
}

func dump(req *http.Request, res *http.Response) {
	const separatorWidth = 60

	reqDump, _ := httputil.DumpRequest(req, true)
	fmt.Printf("\n\n%s", strings.ToUpper("Request Details"))
	fmt.Printf("\n%s\n\n", strings.Repeat("-", separatorWidth))
	fmt.Print(string(reqDump))

	if res != nil {
		respDump, _ := httputil.DumpResponse(res, false)
		fmt.Printf("\n\n%s", strings.ToUpper("Response Details"))
		fmt.Printf("\n%s\n\n", strings.Repeat("-", separatorWidth))
		fmt.Print(string(respDump))
	}
}
