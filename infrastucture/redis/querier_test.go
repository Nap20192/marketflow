package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"marketflow/internal/core/model"
)

func TestQuerier(t *testing.T) {
	client, err := NewClient(context.Background(), RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Fatalf("ailed to connect to redis: %s", err.Error())
	}
	defer client.Close()
	q := NewQueries(client)

	err = q.SaveRawData(context.Background(), "exchanger1", model.Trade{
		Symbol:    "BTCUSDT",
		Price:     30000,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("failed to insert raw data: %s", err.Error())
	}
	s, e, err := q.GetCollection(context.Background())
	if err != nil {
		t.Fatalf("failed to get collection: %s", err.Error())
	}
	fmt.Println(s)
	fmt.Println(e)

	d, err := q.GetRawData(context.Background(), "exchanger1", "BTCUSDT", 30*time.Second)
	if err != nil {
		t.Fatalf("failed to get raw data: %s", err.Error())
	}
	fmt.Println(len(d))
}
