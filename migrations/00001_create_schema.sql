DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
    CREATE DOMAIN UINT256 AS NUMERIC
        CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    ELSE
    ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
    ALTER DOMAIN UINT256 ADD
        CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    END IF;
END $$;


CREATE TABLE IF NOT EXISTS blocks (
    hash  VARCHAR PRIMARY KEY,
    parent_hash VARCHAR  NOT NULL UNIQUE,
    number UINT256 NOT NULL UNIQUE CHECK(number>0),
    timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0),
    rlp_bytes VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS blocks_number ON blocks(number);
CREATE INDEX IF NOT EXISTS blocks_timestamp ON blocks(timestamp);


CREATE TABLE IF NOT EXISTS contracts_event (
    guid  VARCHAR PRIMARY KEY,
    block_hash VARCHAR NOT NULL REFERENCES blocks(hash) ON DELETE CASCADE,
    contract_address VARCHAR NOT NULL,
    transaction_hash VARCHAR NOT NULL,
    log_index     INTEGER NOT NULL,
    event_signature   VARCHAR NOT NULL,
    timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0),
    rlp_bytes VARCHAR NOT NULL
 );
CREATE INDEX IF NOT EXISTS contracts_event_timestamp ON contracts_event(timestamp);
CREATE INDEX IF NOT EXISTS contracts_event_signature ON contracts_event(event_signature);


CREATE TABLE IF NOT EXISTS addresses (
    guid  VARCHAR PRIMARY KEY,
    user_uid  VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    address_type SMALLINT NOT NULL DEFAULT 0,
    private_key VARCHAR NOT NULL,
    public_key VARCHAR NOT NULL,
    balance  VARCHAR NOT NULL,
    timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0)
);
CREATE INDEX IF NOT EXISTS addresses_user_uid ON addresses(user_uid);
CREATE INDEX IF NOT EXISTS addresses_address ON addresses(address);
CREATE INDEX IF NOT EXISTS addresses_timestamp ON addresses(timestamp);


CREATE TABLE IF NOT EXISTS transactions (
     guid VARCHAR PRIMARY KEY,
     block_hash VARCHAR NOT NULL,
     block_number UINT256 NOT NULL UNIQUE CHECK(block_number>0),
     hash VARCHAR NOT NULL,
     from_address VARCHAR NOT NULL,
     to_address VARCHAR NOT NULL,
     fee VARCHAR NOT NULL,
     amount VARCHAR NOT NULL,
     status SMALLINT NOT NULL DEFAULT 0,
     transaction_index UINT256 NOT NULL UNIQUE,
     tx_type SMALLINT NOT NULL DEFAULT 0,
     timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0),
     r VARCHAR NOT NULL,
     s VARCHAR NOT NULL,
     v VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS transactions_timestamp ON transactions(timestamp);


CREATE TABLE IF NOT EXISTS deposit (
    guid  VARCHAR PRIMARY KEY,
    block_hash VARCHAR NOT NULL,
    block_number UINT256 NOT NULL UNIQUE CHECK(block_number>0),
    hash VARCHAR NOT NULL,
    from_address VARCHAR NOT NULL,
    to_address VARCHAR NOT NULL,
    fee VARCHAR NOT NULL,
    amount VARCHAR NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    transaction_index UINT256 NOT NULL UNIQUE,
    timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0),
    r VARCHAR NOT NULL,
    s VARCHAR NOT NULL,
    v VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS deposit_hash ON deposit(hash);
CREATE INDEX IF NOT EXISTS deposit_timestamp ON deposit(timestamp);


CREATE TABLE IF NOT EXISTS withdraw (
    guid  VARCHAR PRIMARY KEY,
    block_hash VARCHAR NOT NULL,
    block_number UINT256 NOT NULL UNIQUE CHECK(block_number>0),
    hash VARCHAR NOT NULL,
    from_address VARCHAR NOT NULL,
    to_address VARCHAR NOT NULL,
    fee VARCHAR NOT NULL,
    amount VARCHAR NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    transaction_index UINT256 NOT NULL UNIQUE,
    timestamp INTEGER NOT NULL UNIQUE CHECK(timestamp>0),
    tx_sign_hex VARCHAR NOT NULL,
    r VARCHAR NOT NULL,
    s VARCHAR NOT NULL,
    v VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS withdraw_hash ON withdraw(hash);
CREATE INDEX IF NOT EXISTS withdraw_timestamp ON withdraw(timestamp);
