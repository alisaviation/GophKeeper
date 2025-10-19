-- Удаляем триггеры и функции
DROP TRIGGER IF EXISTS increment_secrets_version ON secrets;
DROP TRIGGER IF EXISTS update_secrets_updated_at ON secrets;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS increment_user_secrets_version();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы
DROP TABLE IF EXISTS user_secrets_version;
DROP TABLE IF EXISTS secrets;
DROP TABLE IF EXISTS users;

-- Удаляем расширение
DROP EXTENSION IF EXISTS "uuid-ossp";