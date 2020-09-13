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
	"github.com/segmentio/kafka-go"
	"net/url"
	"strings"
	"time"
)

type kafkaResource struct {
	url.URL
}

func (r *kafkaResource) Await(ctx context.Context) error {
	r.normalize()

	conn, err := r.dialer().DialContext(ctx, "tcp", r.URL.Host)
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

func (r *kafkaResource) dialer() *kafka.Dialer {
	return &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS:       r.tlsConfig(),
	}
}

func (r *kafkaResource) tlsConfig() *tls.Config {
	if r.URL.Scheme == "kafkas" {
		return &tls.Config{InsecureSkipVerify: r.skipTLSVerification()}
	}
	return nil
}

func (r *kafkaResource) skipTLSVerification() bool {
	opts := parseFragment(r.URL.Fragment)
	vals, ok := opts["tls"]
	return ok && len(vals) == 1 && vals[0] == "skip-verify"
}

func (r *kafkaResource) normalize() {
	if len(r.URL.Port()) == 0 {
		var defaultPort int
		switch r.URL.Scheme {
		case "kafka":
			defaultPort = 9092
		case "kafkas":
			defaultPort = 9093
		}
		r.URL.Host = fmt.Sprintf("%v:%v", r.URL.Hostname(), defaultPort)
	}
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
