-- Создание таблицы пользователей
CREATE TABLE users (
                       id VARCHAR(36) PRIMARY KEY,
                       login VARCHAR(50) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Создание таблицы секретов
CREATE TABLE secrets (
                         id VARCHAR(36) PRIMARY KEY,
                         user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         type VARCHAR(20) NOT NULL,
                         name VARCHAR(255) NOT NULL,
                         encrypted_data BYTEA NOT NULL,
                         encrypted_meta BYTEA,
                         version BIGINT NOT NULL DEFAULT 1,
                         created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                         updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                         is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- Индексы для улучшения производительности
CREATE INDEX idx_secrets_user_id ON secrets(user_id);
CREATE INDEX idx_secrets_user_type ON secrets(user_id, type);
CREATE INDEX idx_secrets_version ON secrets(user_id, version);
CREATE INDEX idx_secrets_updated ON secrets(updated_at);
CREATE INDEX idx_users_login ON users(login);

-- Таблица для отслеживания версий секретов пользователя
CREATE TABLE user_secrets_version (
                                      user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                                      current_version BIGINT NOT NULL DEFAULT 0,
                                      last_sync_at TIMESTAMP WITH TIME ZONE
);