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
    timestamp INTEGER NOT NULL CHECK(timestamp>0),
    rlp_bytes VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS blocks_number ON blocks(number);
CREATE INDEX IF NOT EXISTS blocks_timestamp ON blocks(timestamp);


CREATE TABLE IF NOT EXISTS tokens (
    guid  VARCHAR PRIMARY KEY,
    token_address VARCHAR NOT NULL,
    unit SMALLINT NOT NULL DEFAULT 18,
    token_name VARCHAR NOT NULL,
    collect_amount  UINT256 NOT NULL CHECK(collect_amount>0),
    timestamp INTEGER NOT NULL CHECK(timestamp>0)
 );
CREATE INDEX IF NOT EXISTS tokens_timestamp ON tokens(timestamp);
CREATE INDEX IF NOT EXISTS tokens_token_address ON tokens(token_address);


CREATE TABLE IF NOT EXISTS addresses (
    guid  VARCHAR PRIMARY KEY,
    user_uid  VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    address_type SMALLINT NOT NULL DEFAULT 0,
    private_key VARCHAR NOT NULL,
    public_key VARCHAR NOT NULL,
    timestamp INTEGER NOT NULL CHECK(timestamp>0)
);
CREATE INDEX IF NOT EXISTS addresses_user_uid ON addresses(user_uid);
CREATE INDEX IF NOT EXISTS addresses_address ON addresses(address);
CREATE INDEX IF NOT EXISTS addresses_timestamp ON addresses(timestamp);


CREATE TABLE IF NOT EXISTS balances (
     guid  VARCHAR PRIMARY KEY,
     address  VARCHAR NOT NULL,
     address_type SMALLINT NOT NULL DEFAULT 0,
     token_address VARCHAR NOT NULL,
     balance  UINT256 NOT NULL CHECK(balance>=0),
     lock_balance  UINT256 NOT NULL,
     timestamp INTEGER NOT NULL CHECK(timestamp>0)
);
CREATE INDEX IF NOT EXISTS balances_address ON balances(address);
CREATE INDEX IF NOT EXISTS balances_timestamp ON balances(timestamp);



CREATE TABLE IF NOT EXISTS transactions (
     guid VARCHAR PRIMARY KEY,
     block_hash VARCHAR NOT NULL,
     block_number UINT256 NOT NULL CHECK(block_number>0),
     hash VARCHAR NOT NULL,
     from_address VARCHAR NOT NULL,
     to_address VARCHAR NOT NULL,
     token_address VARCHAR NOT NULL,
     fee UINT256 NOT NULL,
     amount UINT256 NOT NULL,
     status SMALLINT NOT NULL DEFAULT 0,
     transaction_index UINT256 NOT NULL,
     tx_type SMALLINT NOT NULL DEFAULT 0,
     timestamp INTEGER NOT NULL CHECK(timestamp>0)
);
CREATE INDEX IF NOT EXISTS transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS transactions_timestamp ON transactions(timestamp);


CREATE TABLE IF NOT EXISTS deposits (
    guid  VARCHAR PRIMARY KEY,
    block_hash VARCHAR NOT NULL,
    block_number UINT256 NOT NULL CHECK(block_number>0),
    hash VARCHAR NOT NULL,
    from_address VARCHAR NOT NULL,
    to_address VARCHAR NOT NULL,
    token_address VARCHAR NOT NULL,
    fee UINT256 NOT NULL,
    amount UINT256 NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    transaction_index UINT256 NOT NULL,
    timestamp INTEGER NOT NULL CHECK(timestamp>0)
);
CREATE INDEX IF NOT EXISTS deposits_hash ON deposits(hash);
CREATE INDEX IF NOT EXISTS deposits_timestamp ON deposits(timestamp);


CREATE TABLE IF NOT EXISTS withdraws (
    guid  VARCHAR PRIMARY KEY,
    block_hash VARCHAR NOT NULL,
    block_number UINT256 NOT NULL CHECK(block_number>0),
    hash VARCHAR NOT NULL,
    from_address VARCHAR NOT NULL,
    to_address VARCHAR NOT NULL,
    token_address VARCHAR NOT NULL,
    fee UINT256 NOT NULL,
    amount UINT256 NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    transaction_index UINT256 NOT NULL,
    timestamp INTEGER NOT NULL CHECK(timestamp>0),
    tx_sign_hex VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS withdraws_hash ON withdraws(hash);
CREATE INDEX IF NOT EXISTS withdraws_timestamp ON withdraws(timestamp);
