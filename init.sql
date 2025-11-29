CREATE TABLE market (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    average_price NUMERIC(18, 8) NOT NULL,
    min_price NUMERIC(18, 8) NOT NULL,
    max_price NUMERIC(18, 8) NOT NULL
);

CREATE TABLE raw_data (
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    price NUMERIC(18, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_market_data_pair ON market (pair_name);
