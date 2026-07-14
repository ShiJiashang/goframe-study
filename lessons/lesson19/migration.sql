CREATE TABLE IF NOT EXISTS payment_callbacks (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    trade_no    VARCHAR(64) NOT NULL,
    order_no    VARCHAR(64) NOT NULL,
    status      VARCHAR(20) NOT NULL,
    payload     JSON NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_callbacks_trade_no (trade_no),
    KEY idx_payment_callbacks_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
