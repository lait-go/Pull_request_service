CREATE TABLE orders (
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT NOT NULL,
    entry              TEXT,
    locale             TEXT,
    internal_signature TEXT,
    customer_id        TEXT,
    delivery_service   TEXT,
    shardkey           TEXT,
    sm_id              INT,
    date_created       TIMESTAMP WITH TIME ZONE NOT NULL,
    oof_shard          TEXT
);

CREATE TABLE deliverys (
    id        SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    phone     TEXT NOT NULL,
    zip       TEXT NOT NULL,
    city      TEXT NOT NULL,
    address   TEXT NOT NULL,
    region    TEXT NOT NULL,
    email     TEXT NOT NULL
);

CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    order_uid       TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction     TEXT NOT NULL,
    request_id      TEXT NOT NULL,
    currency        TEXT NOT NULL,
    provider        TEXT NOT NULL,
    amount          INT NOT NULL,
    payment_dt      BIGINT NOT NULL,
    bank            TEXT NOT NULL,
    delivery_cost   INT NOT NULL,
    goods_total     INT NOT NULL,
    custom_fee      INT NOT NULL
);

CREATE TABLE items (
    id            SERIAL PRIMARY KEY,
    order_uid     TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id       BIGINT NOT NULL,
    track_number  TEXT NOT NULL,
    price         INT NOT NULL,
    rid           TEXT NOT NULL,
    name          TEXT NOT NULL,
    sale          INT NOT NULL,
    size          TEXT NOT NULL,
    total_price   INT NOT NULL,
    nm_id         BIGINT NOT NULL,
    brand         TEXT NOT NULL,
    status        INT NOT NULL
);