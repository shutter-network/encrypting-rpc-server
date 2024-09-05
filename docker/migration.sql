CREATE TABLE IF NOT EXISTS transaction_details (
    address VARCHAR(255) NOT NULL,
    nonce BIGINT NOT NULL,
    tx_hash VARCHAR(255) NOT NULL,
    encrypted_tx_hash VARCHAR(255),
    submission_time BIGINT NOT NULL,
    inclusion_time BIGINT,
    is_cancellation BOOLEAN NOT NULL,
    PRIMARY KEY (address, nonce, tx_hash, encrypted_tx_hash)
);
CREATE INDEX IF NOT EXISTS idx_address_nonce on transaction_details (address, nonce);
CREATE INDEX IF NOT EXISTS idx_tx_hash on transaction_details (tx_hash);
CREATE INDEX IF NOT EXISTS idx_encrypted_tx_hash on transaction_details (encrypted_tx_hash);


DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_publication
        WHERE pubname = 'mypub'
    ) THEN
        CREATE PUBLICATION mypub FOR TABLE transaction_details;
    END IF;
END $$;