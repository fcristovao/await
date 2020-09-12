// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTLSSkipVerify(t *testing.T) {
	shutdownServer := setupHttpsServer(t)
	defer shutdownServer()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	resources, _ := parseResources([]string{
		"https://localhost:55372",
		"https://localhost:55372#tls=skip-verify",
	})

	if err := resources[0].Await(ctx); err == nil {
		t.Errorf("Should have failed when verifying TLS, but succeeded.")
	}

	if err := resources[1].Await(ctx); err != nil {
		t.Errorf("Should have skipped TLS verification, but didn't: %v", err)
	}
}

func setupHttpsServer(t *testing.T) func() {
	certFile, keyFile, cleanupCerts := setupTestCertificates(t)

	server := &http.Server{
		Addr: ":55372",
	}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprint(res, "Hello: "+req.Host)
	})

	// The separation of the listening socket from the serving of the http requests
	// is to guarantee one can immediately contact the server on this method's return
	ln, err := net.Listen("tcp", ":55372")
	if err != nil {
		t.Errorf("Unable to setup listening socket for webserver: %v", err)
	}

	go func() {
		_ = server.ServeTLS(ln, certFile, keyFile)
	}()

	return func() {
		_ = server.Close()
		_ = ln.Close()
		cleanupCerts()
	}
}

func setupTestCertificates(t *testing.T) (string, string, func()) {
	certFile := "testCert.crt"
	keyFile := "testCert.key"
	cert, key, err := CertWithKeyPair()
	if err != nil {
		t.Errorf("failed to generate test certificates: %v", err)
	}
	if err := ioutil.WriteFile(certFile, cert, 0644); err != nil {
		t.Errorf("failed to write cert fixture to %s: %v", certFile, err)
	}
	if err := ioutil.WriteFile(keyFile, key, 0644); err != nil {
		t.Errorf("failed to write key fixture to %s: %v", keyFile, err)
	}
	return certFile, keyFile, func() {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
	}
}

func CertWithKeyPair() ([]byte, []byte, error) {
	bits := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	tpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "wronghost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	derCert, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	buf := &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	})
	if err != nil {
		return nil, nil, err
	}

	pemCert := buf.Bytes()

	buf = &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, nil, err
	}
	pemKey := buf.Bytes()

	return pemCert, pemKey, nil
}
