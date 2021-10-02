package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

func NewTLSConfig() (*tls.Config, error) {
	caCert, err := ioutil.ReadFile("/etc/ssl/certs/ca-certificates.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:                  caCertPool,
		InsecureSkipVerify:       true,
	}, nil
}
