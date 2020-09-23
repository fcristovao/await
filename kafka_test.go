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
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

// To run these tests, a kafka cluster needs to be available.
// For local development, at time of writing, this proved enough:
// https://hub.docker.com/r/bitnami/kafka/
//
// Basically, just run:
// $ cd testdata
// $ docker-compose up -d
//
// To stop it, run:
// $ docker-compose down

func TestKafkaResource_FailAwaitWhenKafkaIsNotThere(t *testing.T) {
	port := "55373"
	resource := fmt.Sprintf("kafka://localhost:%v", port)
	ensurePortClosed(t, port)

	if err := resourceAwait(t, resource); err == nil {
		t.Errorf("Should have failed to connect to Kafka, but succeeded.")
	}

	// Setup an HTTP server on the port that we say Kafka is.
	// Allows us to test that await will actually work only when connecting to a Kakfka cluster.
	shutdownServer := setupHttpServer(t, port)
	defer shutdownServer()

	if err := resourceAwait(t, resource); err == nil {
		t.Errorf("Should have failed to connect to Kafka, but succeeded.")
	}
}

func TestKafkaResource_Await(t *testing.T) {
	ensureKafkaAvailable(t)

	if err := resourceAwait(t, "kafka://localhost:9092"); err != nil {
		t.Errorf("Should have connected to Kafka, but failed to: %v.", err)
	}
}

func TestKafkaResource_AwaitForAnyTopic(t *testing.T) {
	conn := ensureKafkaAvailable(t)

	if err := clearAllTopics(conn); err != nil {
		t.Fatalf("Unable to clear topics to execute test: %v.", err)
	}

	if err := resourceAwait(t, "kafka://localhost:9092#topics"); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having no topics.")
	}

	topicConfig := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	if err := conn.CreateTopics(topicConfig); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, "kafka://localhost:9092#topics"); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite having topics: %v.", err)
	}
}

func TestKafkaResource_AwaitForSpecificTopic(t *testing.T) {
	conn := ensureKafkaAvailable(t)

	if err := clearAllTopics(conn); err != nil {
		t.Fatalf("Unable to clear topics to execute test: %v.", err)
	}

	topicConfig1 := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	resource1 := fmt.Sprintf("kafka://localhost:9092#topics=%v", topicConfig1.Topic)
	if err := resourceAwait(t, resource1); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having no topics.")
	}

	if err := conn.CreateTopics(topicConfig1); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, resource1); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite the required topic existing: %v.", err)
	}

	topicConfig2 := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	resource2 := fmt.Sprintf("kafka://localhost:9092#topics=%v,%v", topicConfig1.Topic, topicConfig2.Topic)
	if err := resourceAwait(t, resource2); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having only one topic required.")
	}

	if err := conn.CreateTopics(topicConfig2); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, resource2); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite the required topics existing: %v.", err)
	}
}

func TestKafkaTLSResource_Await(t *testing.T) {
	ensureKafkaTLSAvailable(t)

	// Notice the `kafka` instead of `kafkas`:
	if err := resourceAwait(t, "kafka://localhost:9093"); err == nil {
		t.Errorf("Should have failed to Kafka via TLS when not using `kafkas` scheme, but didn't.")
	}

	if err := resourceAwait(t, "kafkas://localhost:9093#tls=skip-verify"); err != nil {
		t.Errorf("Should have connected to Kafka via TLS and ignore certificate verification, but failed to: %v.", err)
	}
}

func TestKafkaTLSResource_AwaitForAnyTopic(t *testing.T) {
	conn := ensureKafkaTLSAvailable(t)

	if err := clearAllTopics(conn); err != nil {
		t.Fatalf("Unable to clear topics to execute test: %v.", err)
	}

	if err := resourceAwait(t, "kafkas://localhost:9093#topics&tls=skip-verify"); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having no topics.")
	}

	topicConfig := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	if err := conn.CreateTopics(topicConfig); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, "kafkas://localhost:9093#tls=skip-verify&topics"); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite having topics: %v.", err)
	}
}

func TestKafkaTLSResource_AwaitForSpecificTopic(t *testing.T) {
	conn := ensureKafkaTLSAvailable(t)

	if err := clearAllTopics(conn); err != nil {
		t.Fatalf("Unable to clear topics to execute test: %v.", err)
	}

	topicConfig1 := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	resource1 := fmt.Sprintf("kafkas://localhost:9093#tls=skip-verify&topics=%v", topicConfig1.Topic)
	if err := resourceAwait(t, resource1); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having no topics.")
	}

	if err := conn.CreateTopics(topicConfig1); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, resource1); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite the required topic existing: %v.", err)
	}

	topicConfig2 := kafka.TopicConfig{
		Topic:             randomString(20), // Ensure that a random name is used
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	resource2 := fmt.Sprintf("kafkas://localhost:9093#tls=skip-verify&topics=%v,%v", topicConfig1.Topic, topicConfig2.Topic)
	if err := resourceAwait(t, resource2); err == nil {
		t.Errorf("Should not have proceeded, but it did, despite having only one topic required.")
	}

	if err := conn.CreateTopics(topicConfig2); err != nil {
		t.Fatalf("Unable to create topic to proceed with testing: %v.", err)
	}

	if err := resourceAwait(t, resource2); err != nil {
		t.Errorf("Should have proceeded, but didn't, despite the required topics existing: %v.", err)
	}
}

func TestKafkaTLSWithSASLResource_Await(t *testing.T) {
	ensureKafkaTLSWithSASLAvailable(t)

	if err := resourceAwait(t, "kafkas://localhost:9094#tls=skip-verify"); err == nil {
		t.Errorf("Should have failed to Kafka via TLS with SCRAM authentication, but didn't.")
	}

	if err := resourceAwait(t, "kafkas://user:password@localhost:9094#tls=skip-verify"); err != nil {
		t.Errorf("Should have connected to Kafka via TLS with SCRAM authentication, but failed to: %v.", err)
	}
}

func clearAllTopics(conn *kafka.Conn) error {
	topics, err := existingTopics(conn)
	if err != nil {
		return err
	}

	if err := conn.DeleteTopics(topics...); err != nil {
		return err
	}

	newTopics, _ := existingTopics(conn)
	if len(newTopics) != 0 {
		return fmt.Errorf("deletion of all topics still didn't delete all topics. Topics left: %v", newTopics)
	}

	return nil
}

func resourceAwait(t *testing.T, s string) error {
	resource, err := parseResource(s)
	if err != nil {
		t.Fatalf("Failed to parse Resource: %v.", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	return resource.Await(ctx)
}

func ensurePortClosed(t *testing.T, port string) {
	address := net.JoinHostPort("localhost", port)
	if conn, _ := net.DialTimeout("tcp", address, 1*time.Second); conn != nil {
		t.Skipf("Connecting to %v succeded, when we require it to fail to proceed with testing. Skipping", address)
	}
}

func ensureKafkaAvailable(t *testing.T) *kafka.Conn {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	conn, err := kafka.DialContext(ctx, "tcp", "localhost:9092")
	if err != nil {
		t.Skipf("No kafka available for testing (%v), skipping.", err)
	}
	return conn
}

func ensureKafkaTLSAvailable(t *testing.T) *kafka.Conn {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	dialer := &kafka.Dialer{
		TLS: &tls.Config{InsecureSkipVerify: true},
	}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:9093")
	if err != nil {
		t.Skipf("No kafka available for testing (%v), skipping.", err)
	}
	return conn
}

func ensureKafkaTLSWithSASLAvailable(t *testing.T) *kafka.Conn {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	dialer := &kafka.Dialer{
		TLS: &tls.Config{InsecureSkipVerify: true},
	}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:9094")
	if err != nil {
		t.Skipf("No kafka available for testing (%v), skipping.", err)
	}
	return conn
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
