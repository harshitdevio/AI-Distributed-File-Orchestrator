package network

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"organiser/classifier"
	"organiser/server"

	"github.com/nats-io/nats.go"
)


func DispatchClassificationRequest(files []classifier.FileInfo) ([]server.ClassificationResult, error) {
	// Resolve network location
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	// Establish connection boundary
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("network connection failure: %w", err)
	}
	defer nc.Close()

	// 3. Serialize data transfer object
	reqBytes, err := json.Marshal(files)
	if err != nil {
		return nil, fmt.Errorf("payload serialization error: %w", err)
	}

	// Execute message exchange over the network fabric (5 minute max execution window)
	replyMsg, err := nc.Request("llm.classify.batch", reqBytes, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("message broker transport timeout/error: %w", err)
	}

	// Unmarshal protocol representation back to domain types
	var finalResults []server.ClassificationResult
	if err := json.Unmarshal(replyMsg.Data, &finalResults); err != nil {
		return nil, fmt.Errorf("payload deserialization error: %w", err)
	}

	return finalResults, nil
}