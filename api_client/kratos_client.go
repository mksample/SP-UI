package api_client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sync"

	"github.com/davidoram/kratos-selfservice-ui-go/options"
	kratos "github.com/ory/kratos-client-go"
)

var (
	publicClientInstance *kratos.APIClient
)

func InitPublicClient(opt *options.Options) (*kratos.APIClient, error) {
	var tlsConfig *tls.Config
	var err error
	if opt.TLSCertPath != "" {
		tlsConfig, err = NewTLSConfig(opt.TLSCertPath, opt.TLSKeyPath, opt.TLSCaPath)
		if err != nil {
			return nil, err
		}
	}
	url := opt.KratosPublicURL
	configuration := kratos.NewConfiguration()
	configuration.UserAgent = "Public self service UI"
	configuration.Host = url.Host
	configuration.Scheme = url.Scheme
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		configuration.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Jar: cj,
		}
	}
	configuration.Servers = []kratos.ServerConfiguration{
		{
			URL: url.Path,
		},
	}
	publicClientInstance = kratos.NewAPIClient(configuration)

	return publicClientInstance, nil
}

func PublicClient() *kratos.APIClient {
	return publicClientInstance
}

var (
	adminClientInstance *kratos.APIClient
)

func InitAdminClient(opt *options.Options) (*kratos.APIClient, error) {
	var tlsConfig *tls.Config
	var err error
	if opt.TLSCertPath != "" {
		tlsConfig, err = NewTLSConfig(opt.TLSCertPath, opt.TLSKeyPath, opt.TLSCaPath)
		if err != nil {
			return nil, err
		}
	}

	url := opt.KratosAdminURL
	configuration := kratos.NewConfiguration()
	configuration.UserAgent = "Admin self service UI"
	configuration.Host = url.Host
	configuration.Scheme = url.Scheme
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		configuration.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Jar: cj,
		}
	}
	configuration.Servers = []kratos.ServerConfiguration{
		{
			URL: url.Path,
		},
	}
	adminClientInstance = kratos.NewAPIClient(configuration)

	return adminClientInstance, nil
}

func AdminClient() *kratos.APIClient {
	return adminClientInstance
}

var (
	whoamiClientOnce     sync.Once
	whoamiClientInstance *kratos.APIClient
)

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		log.Printf("here")
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		log.Printf("no here")
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}
