package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"marketflow/internal/adapters/secondary/storage"
	"marketflow/internal/core/model"
)

var ErrNoRows = fmt.Errorf("no data found")

const getAverage = `
SELECT AVG(average_price) AS avg_price
FROM market
WHERE
    pair_name = $1
    AND exchange = $2
    AND timestamp > now() - ($3 * interval '1 second')
`

func (q *Queries) GetAverage(ctx context.Context, arg storage.Params) (float64, error) {
	seconds := int64(arg.Interval.Seconds())
	row := q.db.QueryRow(ctx, getAverage, arg.PairName, arg.Exchange, seconds)
	var avgPrice sql.NullFloat64
	err := row.Scan(&avgPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, err
	}
	if !avgPrice.Valid {
		return 0, ErrNoRows
	}

	return avgPrice.Float64, nil
}

const getMax = `
SELECT MAX(max_price) AS max_price
FROM market
WHERE
    pair_name = $1
    AND exchange = $2
    AND timestamp > now() - ($3 * interval '1 second')
`

func (q *Queries) GetMax(ctx context.Context, arg storage.Params) (float64, error) {
	seconds := int64(arg.Interval.Seconds())
	row := q.db.QueryRow(ctx, getMax, arg.PairName, arg.Exchange, seconds)
	var maxPrice sql.NullFloat64
	err := row.Scan(&maxPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, err
	}
	if !maxPrice.Valid {
		return 0, ErrNoRows
	}

	return maxPrice.Float64, nil
}

const getMin = `
SELECT MIN(min_price) AS min_price
FROM market
WHERE
    pair_name = $1
    AND exchange = $2
    AND timestamp > now() - ($3 * interval '1 second')
`

func (q *Queries) GetMin(ctx context.Context, arg storage.Params) (float64, error) {
	seconds := int64(arg.Interval.Seconds())
	row := q.db.QueryRow(ctx, getMin, arg.PairName, arg.Exchange, seconds)
	var minPrice sql.NullFloat64
	err := row.Scan(&minPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, err
	}
	if !minPrice.Valid {
		return 0, ErrNoRows
	}

	return minPrice.Float64, nil
}

const insertMarket = `
INSERT INTO
    market (
        pair_name,
        exchange,
        average_price,
        min_price,
        max_price
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    id, pair_name, exchange, timestamp, average_price, min_price, max_price
`

func (q *Queries) InsertMarket(ctx context.Context, arg storage.InsertMarketParams) (model.AggregatedData, error) {
	row := q.db.QueryRow(ctx, insertMarket,
		arg.PairName,
		arg.Exchange,
		arg.AveragePrice,
		arg.MinPrice,
		arg.MaxPrice,
	)
	var data model.AggregatedData

	err := row.Scan(
		&data.ID,
		&data.PairName,
		&data.Exchange,
		&data.Timestamp,
		&data.AveragePrice,
		&data.MinPrice,
		&data.MaxPrice,
	)
	return data, err
}
