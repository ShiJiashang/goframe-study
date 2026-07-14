CREATE DATABASE IF NOT EXISTS goframe_mall
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE goframe_mall;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  username VARCHAR(64) NOT NULL COMMENT '登录名',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1正常，0禁用',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB COMMENT='用户表';

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  name VARCHAR(64) NOT NULL COMMENT '分类名称',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用，0停用',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_categories_name (name)
) ENGINE=InnoDB COMMENT='商品分类表';

CREATE TABLE IF NOT EXISTS products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
  category_id BIGINT UNSIGNED NOT NULL COMMENT '分类ID',
  name VARCHAR(128) NOT NULL COMMENT '商品名称',
  price_cent BIGINT UNSIGNED NOT NULL COMMENT '价格，单位为分',
  stock INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '库存',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1上架，0下架',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_products_category_status (category_id, status),
  KEY idx_products_name (name),
  CONSTRAINT fk_products_category
    FOREIGN KEY (category_id) REFERENCES categories (id)
) ENGINE=InnoDB COMMENT='商品表';

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  order_no VARCHAR(64) NOT NULL COMMENT '订单号',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  total_cent BIGINT UNSIGNED NOT NULL COMMENT '订单总金额，单位为分',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1待支付，2已支付，3已取消',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_no (order_no),
  KEY idx_orders_user_created (user_id, created_at),
  CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB COMMENT='订单表';

CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单明细ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  product_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
  product_name VARCHAR(128) NOT NULL COMMENT '下单时商品名称快照',
  price_cent BIGINT UNSIGNED NOT NULL COMMENT '下单时单价，单位为分',
  quantity INT UNSIGNED NOT NULL COMMENT '购买数量',
  subtotal_cent BIGINT UNSIGNED NOT NULL COMMENT '小计，单位为分',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_order_items_order (order_id),
  KEY idx_order_items_product (product_id),
  CONSTRAINT fk_order_items_order
    FOREIGN KEY (order_id) REFERENCES orders (id),
  CONSTRAINT fk_order_items_product
    FOREIGN KEY (product_id) REFERENCES products (id)
) ENGINE=InnoDB COMMENT='订单明细表';

INSERT INTO users (id, username, password_hash)
VALUES (1, 'demo', 'lesson-only-placeholder')
ON DUPLICATE KEY UPDATE username = VALUES(username);

INSERT INTO categories (id, name, sort)
VALUES (1, '学习用品', 10)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort = VALUES(sort);

INSERT INTO products (id, category_id, name, price_cent, stock, status)
VALUES
  (1, 1, 'GoFrame 实战手册', 5990, 10, 1),
  (2, 1, '机械键盘', 29900, 20, 1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  price_cent = VALUES(price_cent),
  stock = VALUES(stock),
  status = VALUES(status);
