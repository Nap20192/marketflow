package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"marketflow/internal/core/model"
)

func TestQueries_Fallback(t *testing.T) {
	db, err := NewClient(PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "marketflow",
		Password: "marketflow",
		DBName:   "marketflow",
	})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %s", err.Error())
	}
	defer db.Close()
	q := New(db.Pool)

	e, s, err := q.GetCollection(context.Background())
	if err != nil {
		t.Fatalf("failed to get collection: %s", err.Error())
	}
	fmt.Println(e)
	fmt.Println(s)
	err = q.SaveRawData(context.Background(), "exchanger1", model.Trade{
		Symbol:    "BTCUSDT",
		Price:     30000,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("failed to insert raw data: %s", err.Error())
	}
}
