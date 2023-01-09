package api_client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"

	"github.com/davidoram/kratos-selfservice-ui-go/options"
	kratos "github.com/ory/kratos-client-go"
)

// Kratos client instances
var (
	publicClientInstance *kratos.APIClient
	adminClientInstance  *kratos.APIClient
)

// Gets the public client
func PublicClient() *kratos.APIClient {
	return publicClientInstance
}

// Gets the admin client
func AdminClient() *kratos.APIClient {
	return adminClientInstance
}

// Initializes the public client
func InitPublicClient(opt *options.Options) (*kratos.APIClient, error) {
	cfg, err := NewKratosConfig(opt)
	if err != nil {
		return nil, err
	}

	publicClientInstance = kratos.NewAPIClient(cfg)

	return publicClientInstance, nil
}

// Initializes the admin client
func InitAdminClient(opt *options.Options) (*kratos.APIClient, error) {
	cfg, err := NewKratosConfig(opt)
	if err != nil {
		return nil, err
	}

	adminClientInstance = kratos.NewAPIClient(cfg)

	return adminClientInstance, nil
}

// Creates a kratos client config from options
func NewKratosConfig(opt *options.Options) (cfg *kratos.Configuration, err error) {
	url := opt.KratosPublicURL
	cfg = kratos.NewConfiguration()

	cfg.Host = url.Host
	cfg.Scheme = url.Scheme
	cfg.Servers = []kratos.ServerConfiguration{{URL: url.Path}}
	cfg.UserAgent = "Public self service UI"
	cj, err := cookiejar.New(nil) // TODO: don't know if this is actually necessary
	if err != nil {
		return nil, err
	}
	cfg.HTTPClient = &http.Client{Jar: cj}

	if opt.TLSCertPath != "" {
		tlsConfig, err := NewTLSConfig(opt.TLSCertPath, opt.TLSKeyPath, opt.TLSCaPath)
		if err != nil {
			return nil, err
		}
		cfg.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	return cfg, nil
}

// Creates a TLS config from certificate/key paths
func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	cfg := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		log.Printf("here")
		return &cfg, err
	}
	cfg.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		log.Printf("no here")
		return &cfg, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cfg.RootCAs = caCertPool

	cfg.BuildNameToCertificate()
	return &cfg, err
}
