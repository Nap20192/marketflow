package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"marketflow/internal/core/model"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNoConnection = fmt.Errorf("no connection to redis server")
	ErrNoData       = errors.New("no data")
	ErrParse        = errors.New("parse error")
)

func mapRedisErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrNoConnection) {
		return ErrNoConnection
	}

	if errors.Is(err, redis.Nil) {
		return err
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "dial tcp") || strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("%w: %v", ErrNoConnection, err)
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Err != nil && strings.Contains(strings.ToLower(opErr.Err.Error()), "connection refused") {
			return fmt.Errorf("%w: %v", ErrNoConnection, err)
		}
	}

	return err
}

func (q *Queries) SaveRawData(ctx context.Context, exchanger string, data model.Trade) error {
	key := fmt.Sprintf("prices:%s:%s", exchanger, data.Symbol)

	ts := float64(time.Now().Unix())

	pipe := q.client.TxPipeline()

	pipe.ZAdd(ctx, key, redis.Z{Score: float64(ts), Member: fmt.Sprintf(`{"price":%.8f,"ts":%d}`, data.Price, int64(ts))})
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprint(float64(ts-120)))
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return mapRedisErr(fmt.Errorf("save raw data %s:%s: %w", exchanger, data.Symbol, err))
	}

	return nil
}

func (q *Queries) GetRawData(ctx context.Context, exchanger string, symbol string, interval time.Duration) ([]model.Trade, error) {
	key := fmt.Sprintf("prices:%s:%s", exchanger, symbol)

	now := time.Now().Unix()
	from := now - int64(interval.Seconds())

	res, err := q.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: fmt.Sprint(from), Max: fmt.Sprint(now)}).Result()
	if err != nil {
		return nil, mapRedisErr(fmt.Errorf("redis ZRangeByScore %s: %w", key, err))
	}

	trades := make([]model.Trade, 0, len(res))
	for _, r := range res {
		var trade model.Trade
		_, err := fmt.Sscanf(r.Member.(string), `{"price":%f,"ts":%d}`, &trade.Price, &trade.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse member %q: %v: %w", r.Member, err, ErrParse)
		}

		trade.Symbol = symbol
		trades = append(trades, trade)
	}
	return trades, nil
}

func (q *Queries) GetCollection(ctx context.Context) ([]string, []string, error) {
	pattern := "prices:*"
	exchangers := make(map[string]struct{})
	symbols := make(map[string]struct{})

	resExchangers := make([]string, 0)
	resSymbols := make([]string, 0)
	var cursor uint64
	for {
		keys, nextCursor, err := q.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, nil, mapRedisErr(fmt.Errorf("scan keys: %w", err))
		}
		cursor = nextCursor

		for _, key := range keys {
			var exchanger, symbol string
			parts := strings.SplitN(key, ":", 3)

			if len(parts) != 3 || parts[0] != "prices" {
				slog.Warn("invalid key format", "key", key)
				continue
			}

			exchanger = parts[1]
			symbol = parts[2]

			if _, ok := exchangers[exchanger]; !ok {
				exchangers[exchanger] = struct{}{}
				resExchangers = append(resExchangers, exchanger)
			}

			if _, ok := symbols[symbol]; !ok {
				symbols[symbol] = struct{}{}
				resSymbols = append(resSymbols, symbol)
			}
		}
		if cursor == 0 {
			return resExchangers, resSymbols, nil
		}
	}
}

func (q *Queries) GetLatest(ctx context.Context, exchange string, symbol string) (float64, error) {
	key := fmt.Sprintf("prices:%s:%s", exchange, symbol)

	res, err := q.client.ZRevRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		return 0, mapRedisErr(fmt.Errorf("redis ZRevRange %s: %w", key, err))
	}

	if len(res) == 0 {
		return 0, ErrNoData
	}

	var trade model.Trade

	_, err = fmt.Sscanf(res[0].Member.(string), `{"price":%f,"ts":%d}`, &trade.Price, &trade.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("parse latest %s:%s: %v: %w", exchange, symbol, err, ErrParse)
	}

	return trade.Price, nil
}
