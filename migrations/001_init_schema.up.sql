-- Создаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       login VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индекс для быстрого поиска по логину
CREATE INDEX idx_users_login ON users(login);

-- Таблица секретов
CREATE TABLE secrets (
                         id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                         user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         type VARCHAR(50) NOT NULL,
                         name VARCHAR(500) NOT NULL,
                         encrypted_data BYTEA NOT NULL,
                         encrypted_meta BYTEA,
                         version BIGINT NOT NULL DEFAULT 1,
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                         updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                         is_deleted BOOLEAN DEFAULT FALSE,

    -- Ограничения
                         CHECK (type IN ('login_password', 'text_data', 'binary_data', 'bank_card'))
);

-- Индексы для секретов
CREATE INDEX idx_secrets_user_id ON secrets(user_id);
CREATE INDEX idx_secrets_user_id_type ON secrets(user_id, type) WHERE NOT is_deleted;
CREATE INDEX idx_secrets_user_id_version ON secrets(user_id, version) WHERE NOT is_deleted;
CREATE INDEX idx_secrets_updated_at ON secrets(updated_at) WHERE NOT is_deleted;

-- Таблица для отслеживания версий секретов пользователя
CREATE TABLE user_secrets_version (
                                      user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                                      current_version BIGINT NOT NULL DEFAULT 1,
                                      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_secrets_updated_at BEFORE UPDATE ON secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Функция для увеличения версии секретов пользователя
CREATE OR REPLACE FUNCTION increment_user_secrets_version()
RETURNS TRIGGER AS $$
BEGIN
INSERT INTO user_secrets_version (user_id, current_version)
VALUES (NEW.user_id, 1)
    ON CONFLICT (user_id)
    DO UPDATE SET
    current_version = user_secrets_version.current_version + 1,
               updated_at = NOW();

RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для увеличения версии при изменении секретов
CREATE TRIGGER increment_secrets_version AFTER INSERT OR UPDATE OR DELETE ON secrets
    FOR EACH ROW EXECUTE FUNCTION increment_user_secrets_version();