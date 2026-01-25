package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type EventResponse struct {
	Status    string `json:"status"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Event     Event  `json:"event"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type createReq struct {
	Type    string          `json:"type"`   // movie | user | payment
	Payload json.RawMessage `json:"payload"` // соответствует схеме
}

func main() {
	port := getenv("PORT", "8082")
	brokers := getenv("KAFKA_BROKERS", "kafka:9092")

	// topics соответствуют твоему KAFKA_CREATE_TOPICS
	topics := map[string]string{
		"movie":   "movie-events",
		"user":    "user-events",
		"payment": "payment-events",
	}

	// Producer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 50 * time.Millisecond,
	}
	defer writer.Close()

	// Consumers (в фоне)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for evType, topic := range topics {
		go consumeLoop(ctx, brokers, topic, "events-service-"+evType)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Универсальный endpoint: POST /events
	// body: { "type": "movie|user|payment", "payload": {...} }
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
			return
		}

		var req createReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}

		req.Type = strings.ToLower(strings.TrimSpace(req.Type))
		topic, ok := topics[req.Type]
		if !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: `type must be one of: "movie", "user", "payment"`})
			return
		}
		if len(req.Payload) == 0 || string(req.Payload) == "null" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "payload is required"})
			return
		}

		ev := Event{
			ID:        makeEventID(req.Type),
			Type:      req.Type,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   req.Payload,
		}

		value, err := json.Marshal(ev)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to marshal event"})
			return
		}

		msg := kafka.Message{
			Topic: topic,
			Key:   []byte(ev.ID),
			Value: value,
		}

		// пишем в Kafka
		if err := writer.WriteMessages(r.Context(), msg); err != nil {
			log.Printf("[producer] write failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "kafka write failed"})
			return
		}

		// kafka-go после WriteMessages обычно проставляет Partition/Offset в msg
		resp := EventResponse{
			Status:    "success",
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Event:     ev,
		}

		log.Printf("[producer] produced %s to %s partition=%d offset=%d id=%s",
			ev.Type, topic, msg.Partition, msg.Offset, ev.ID)

		writeJSON(w, http.StatusOK, resp)
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("events-service listening on :%s (brokers=%s)", port, brokers)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// ждём сигнал остановки
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	cancel()
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	_ = server.Shutdown(ctxTimeout)
	log.Println("bye")
}

func consumeLoop(ctx context.Context, brokers, topic, groupID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokers, ","),
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1e3,
		MaxBytes:       1e6,
		CommitInterval: 1 * time.Second,
	})
	defer reader.Close()

	log.Printf("[consumer] started topic=%s group=%s", topic, groupID)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("[consumer] stop topic=%s", topic)
				return
			}
			log.Printf("[consumer] read error topic=%s: %v", topic, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// попробуем распарсить как Event, иначе просто залогируем raw
		var ev Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("[consumer] topic=%s partition=%d offset=%d raw=%s",
				topic, m.Partition, m.Offset, string(m.Value))
			continue
		}

		log.Printf("[consumer] processed type=%s id=%s topic=%s partition=%d offset=%d payload=%s",
			ev.Type, ev.ID, topic, m.Partition, m.Offset, string(ev.Payload))
	}
}

func makeEventID(kind string) string {
	// короткий случайный id: movie-<hex>
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", kind, hex.EncodeToString(b))
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[http] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
