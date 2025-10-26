# GophKeeper
Менеджер паролей GophKeeper

## 🔧 Сборка из исходников


```bash
# Установить зависимости
make deps

# Собрать для текущей платформы
make build

# Собрать для всех платформ (Windows, Linux, macOS)
make build-all

# Установить в систему (Linux/macOS)
make install

# Запустить тесты
make test

# Показать все доступные команды
make help
```

Основные команды

#### Регистрация
```
gophkeeper auth register myusername
```
####  Вход
```
gophkeeper auth login myusername
```
####  Создание секрета
```
gophkeeper secrets create-login "Google" "myemail@gmail.com" "mypassword" --website "https://google.com"
```
####  Синхронизация
```
gophkeeper sync
```

####  Просмотр секретов
```
gophkeeper secrets list
```