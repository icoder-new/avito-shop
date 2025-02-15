CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255)       NOT NULL,
    coins         BIGINT             NOT NULL DEFAULT 1000,
    created_at    TIMESTAMP WITH TIME ZONE    DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS merch
(
    id    BIGSERIAL PRIMARY KEY,
    name  VARCHAR(50) UNIQUE NOT NULL,
    price BIGINT             NOT NULL CHECK (price > 0)
);

CREATE TABLE IF NOT EXISTS user_inventory
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users (id) ON DELETE CASCADE,
    merch_id   BIGINT REFERENCES merch (id) ON DELETE CASCADE,
    quantity   BIGINT NOT NULL          DEFAULT 0 CHECK (quantity >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, merch_id)
);

CREATE TABLE IF NOT EXISTS transactions
(
    id           BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
    to_user_id   BIGINT REFERENCES users (id) ON DELETE CASCADE,
    amount       BIGINT      NOT NULL CHECK (amount > 0),
    type         VARCHAR(20) NOT NULL CHECK (type IN ('transfer', 'purchase')),
    merch_id     BIGINT REFERENCES merch (id) ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_user_inventory_user_id ON user_inventory (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_from_user_id ON transactions (from_user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_user_id ON transactions (to_user_id);

INSERT INTO merch (name, price)
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500)
ON CONFLICT (name) DO NOTHING;

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();