package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"marketflow/internal/core/model"
)

const saveRawData = `
INSERT INTO
	raw_data (
		exchange,
		pair_name,
		price
	)
VALUES ($1, $2, $3);
`

func (q *Queries) SaveRawData(ctx context.Context, exchanger string, data model.Trade) error {
	_, err := q.db.Exec(ctx, saveRawData,
		exchanger,
		data.Symbol,
		data.Price,
	)
	if err != nil {
		return err
	}
	_, _ = q.db.Exec(ctx, "DELETE FROM raw_data WHERE created_at < now() - $1::interval", 1*time.Minute)

	return err
}

const getRawDataByRange = `
SELECT pair_name, price, created_at
FROM raw_data
WHERE pair_name = $1
  AND exchange = $2
  AND created_at > now() - $3::interval
ORDER BY created_at ASC;
`

const deleteOldRawData = `
DELETE FROM raw_data
WHERE created_at < now() - $1::interval;
`

func (q *Queries) GetRawData(ctx context.Context, exchange string, symbol string, interval time.Duration) ([]model.Trade, error) {
	rows, err := q.db.Query(ctx, getRawDataByRange, symbol, exchange, interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trades := make([]model.Trade, 0, len(rows.RawValues()))

	for rows.Next() {
		var trade model.Trade

		var t time.Time
		err := rows.Scan(
			&trade.Symbol,
			&trade.Price,
			&t,
		)
		if err != nil {
			return nil, err
		}

		trade.Timestamp = t.Unix()

		trades = append(trades, trade)
	}
	_, _ = q.db.Exec(ctx, deleteOldRawData, 90*time.Second)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return trades, nil
}

func (q *Queries) GetCollection(ctx context.Context) ([]string, []string, error) {
	rows, err := q.db.Query(ctx, `SELECT DISTINCT exchange FROM raw_data`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	exchangers := make([]string, 0)
	for rows.Next() {
		var exchange string
		if err := rows.Scan(&exchange); err != nil {
			return nil, nil, err
		}
		exchangers = append(exchangers, exchange)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	symbolRows, err := q.db.Query(ctx, `SELECT DISTINCT pair_name FROM raw_data`)
	if err != nil {
		return nil, nil, err
	}
	defer symbolRows.Close()

	symbols := make([]string, 0)
	for symbolRows.Next() {
		var symbol string
		if err := symbolRows.Scan(&symbol); err != nil {
			return nil, nil, err
		}
		symbols = append(symbols, symbol)
	}

	if err := symbolRows.Err(); err != nil {
		return nil, nil, err
	}

	return exchangers, symbols, nil
}

const getLatest = `
SELECT  price
FROM raw_data
WHERE pair_name = $1
  AND exchange = $2
ORDER BY created_at DESC
LIMIT 1;
`

func (q *Queries) GetLatest(ctx context.Context, exchange string, symbol string) (float64, error) {
	row := q.db.QueryRow(ctx, getLatest, symbol, exchange)
	var price float64
	err := row.Scan(&price)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, fmt.Errorf("get latest %s:%s: %w", exchange, symbol, err)
	}
	return price, nil
}
