CREATE TABLE transaction_details (
    address VARCHAR(255) NOT NULL,
    nonce BIGINT NOT NULL,
    tx_hash VARCHAR(255) NOT NULL,
    encrypted_tx_hash VARCHAR(255) NOT NULL,
    submission_time BIGINT NOT NULL,
    inclusion_time BIGINT NOT NULL,
    is_cancellation BOOLEAN NOT NULL,
    PRIMARY KEY (address, nonce, tx_hash)
);
CREATE INDEX idx_address_nonce on transaction_details (address, nonce);
CREATE INDEX idx_tx_hash on transaction_details (tx_hash);
CREATE INDEX idx_encrypted_tx_hash on transaction_details (encrypted_tx_hash);


CREATE PUBLICATION mypub FOR TABLE transaction_details;