-- lesson18 seed：为 users 表准备两个测试账号
--
-- 密码：
--   admin / admin123
--   demo  / demo123
--
-- 哈希由 bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost) 生成
--
-- 使用方法：
--   docker exec -i goframe-mysql mysql -uroot -p12345678 goframe_mall \
--       < lessons/lesson18/seed.sql

USE goframe_mall;

-- 用 username 作为唯一索引匹配。id 交给数据库自动分配，避免和 lesson11 的
-- 现有种子数据（id=1 已被占用）冲突。
INSERT INTO users (username, password_hash, status)
VALUES
  ('admin', '$2a$10$JhfRKwWSr/yOs6wmL9wO6eiObcxjgTjsrxywuUUcktQZZpmHVGgMK', 1)
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  status = VALUES(status);

INSERT INTO users (username, password_hash, status)
VALUES
  ('demo',  '$2a$10$1leThsg7jJeO.6HT50yYAOWTby1qwmf9UUW/.JBbvapdfYKBJxAVC', 1)
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  status = VALUES(status);
