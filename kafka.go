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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type kafkaResource struct {
	url.URL
}

// Ensures that misconfigured Kafka Resources are not possible
func newKafkaResource(u url.URL) (resource, error) {
	if len(u.Port()) == 0 {
		var defaultPort int
		switch u.Scheme {
		case "kafka":
			defaultPort = 9092
		case "kafkas":
			defaultPort = 9093
		}
		u.Host = fmt.Sprintf("%v:%v", u.Hostname(), defaultPort)
	}
	mechanism := getOptOrDefault(u, "sasl", "")
	switch mechanism {
	case "plain", "scram-sha-256", "scram-sha-512", "":
		break
	default:
		return nil, &resourceConfigError{
			Reason: fmt.Errorf("%v: unknown value for 'sasl' configuration: %v", u.String(), mechanism),
		}
	}
	tls := getOptOrDefault(u, "tls", "")
	switch tls {
	case "skip-verify", "":
		break
	default:
		return nil, &resourceConfigError{
			Reason: fmt.Errorf("%v: unknown value for 'tls' configuration: %v", u.String(), tls),
		}
	}
	return &kafkaResource{u}, nil
}

func (r *kafkaResource) Await(ctx context.Context) error {
	conn, err := r.conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// Just opening a connection isn't enough (as any open port would succeed)
	// So ask for the brokers to ensure we are talking with a Kafka cluster:
	if _, err := conn.Brokers(); err != nil {
		return &unavailabilityError{err}
	}

	if shouldWait, topics := r.shouldWaitForTopics(); shouldWait {
		if err := awaitTopics(conn, topics); err != nil {
			return err
		}
	}

	return nil
}

func (r *kafkaResource) conn(ctx context.Context) (*kafka.Conn, error) {
	dialer, err := r.dialer()
	if err != nil {
		return nil, err
	}
	conn, err := dialer.DialContext(ctx, "tcp", r.URL.Host)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *kafkaResource) dialer() (*kafka.Dialer, error) {
	mechanism, err := r.saslMechanism()
	if err != nil {
		return nil, err
	}
	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           r.tlsConfig(),
		SASLMechanism: mechanism,
	}, nil
}

func (r *kafkaResource) tlsConfig() *tls.Config {
	if r.URL.Scheme == "kafkas" {
		return &tls.Config{InsecureSkipVerify: r.skipTLSVerification()}
	}
	return nil
}

func (r *kafkaResource) saslMechanism() (sasl.Mechanism, error) {
	userInfo := r.URL.User
	if userInfo != nil {
		password, _ := userInfo.Password()
		username := userInfo.Username()
		mechanism := r.getOptOrDefault("sasl", "plain")
		switch mechanism {
		case "scram-sha-256":
			return scram.Mechanism(scram.SHA256, username, password)
		case "scram-sha-512":
			return scram.Mechanism(scram.SHA512, username, password)
		case "plain":
			fallthrough
		default:
			return plain.Mechanism{Username: username, Password: password}, nil
		}
	}
	return nil, nil
}

func (r *kafkaResource) getOptOrDefault(key string, defaultVal string) string {
	return getOptOrDefault(r.URL, key, defaultVal)
}

func (r *kafkaResource) skipTLSVerification() bool {
	return r.getOptOrDefault("tls", "verify") == "skip-verify"
}

func (r *kafkaResource) shouldWaitForTopics() (bool, []string) {
	opts := parseFragment(r.URL.Fragment)
	if val, ok := opts["topics"]; ok {
		var topics []string
		if len(val) > 0 && val[0] != "" {
			topics = strings.Split(val[0], ",")
		}
		return true, unique(topics)
	}
	return false, nil
}

func awaitTopics(conn *kafka.Conn, requiredTopics []string) error {
	if len(requiredTopics) == 0 {
		// Check only for the existence of _any_ topic
		partitions, err := conn.ReadPartitions()
		if err != nil {
			return err
		}

		if len(partitions) == 0 {
			return &unavailabilityError{errors.New("no topics found")}
		}

		return nil
	}

	existingTopics, err := existingTopics(conn)
	if err != nil {
		return err
	}

	if foundAll, missingTopics := containsAll(existingTopics, requiredTopics); !foundAll {
		return &unavailabilityError{fmt.Errorf("missing topics: %v", missingTopics)}
	}

	return nil
}

func existingTopics(conn *kafka.Conn) ([]string, error) {
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, err
	}

	var topics []string

	for _, partition := range partitions {
		topics = append(topics, partition.Topic)
	}

	return unique(topics), nil
}
