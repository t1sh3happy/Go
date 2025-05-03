# GoBankAPI

GoBankAPI — это REST API для банковского сервиса, разработанный на языке Go с акцентом на безопасность, расширяемость и работу с реальными банковскими сценариями.

## Возможности

- Регистрация и вход с использованием JWT
- Управление банковскими счетами
- Пополнение, снятие, переводы между счетами
- Виртуальные карты с алгоритмом Луна
- Шифрование PGP + HMAC для данных карт
- Кредитование с аннуитетными платежами
- Шедулер платежей и начисление штрафов
- Интеграция:
  - SMTP (Mailtrap)
  - SOAP-запрос к банку "ЦБ"
- Финансовая аналитика и прогнозирование баланса
- HTML-формы для тестирования (без Postman)

## Как запустить

1. **Установите Go 1.22+**
2. **Создайте `.env` файл** на основе:
```
PORT=8080
JWT_SECRET=supersecretkey
DB_DSN=postgres://postgres:postgres@localhost:5432/gobank?sslmode=disable
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your@mail.com
SMTP_LOGIN=your_login
SMTP_PASS=your_pass
RUN_SCHEDULER=true
```
Для тестирования мною был использован сервис **MailTrap** и мои your_login и your_pass.
**Вы можете использовать свои параметры для тестирования.**


3. **Создайте базу данных `gobank` в PostgreSQL**  
Если вы не используете систему миграций, создайте необходимые таблицы вручную:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email           TEXT UNIQUE NOT NULL,
    username        TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE accounts (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    INT REFERENCES users(id),
    number     TEXT NOT NULL,
    balance    NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transactions (
    id             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    from_account_id INT,
    to_account_id   INT,
    amount          NUMERIC(15,2),
    type            TEXT,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE cards (
    id               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          INT REFERENCES users(id),
    account_id       INT REFERENCES accounts(id),
    number_hmac      TEXT NOT NULL,
    encrypted_data   BYTEA NOT NULL,
    cvv_hash         TEXT NOT NULL,
    created_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE credits (
    id              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         INT REFERENCES users(id),
    account_id      INT REFERENCES accounts(id),
    amount          NUMERIC(15,2) NOT NULL,
    interest_rate   NUMERIC(5,2) NOT NULL,
    months          INT NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE payment_schedules (
    id           INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    credit_id    INT REFERENCES credits(id),
    due_date     DATE,
    amount       NUMERIC(15,2),
    paid         BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMP DEFAULT NOW()
);
```

4. **Запустите сервер:**

```bash
go run ./cmd/main.go
```
Сервер поднимется на http://localhost:8080.

## Тестирование

Все ключевые функции можно протестировать через:

* HTML-формы (ручное тестирование через браузер(уже добавлены в проект))
* Эндпоинты (через Postman или curl при необходимости)

### Публичные маршруты

| Метод | Путь      | Описание       |
| ----- | --------- | -------------- |
| POST  | /register | Регистрация    |
| POST  | /login    | Аутентификация |
| GET   | /ping     | Health-check   |

### Защищённые маршруты (/api)

| Метод | Путь                       | Описание                    |
| ----- | -------------------------- | --------------------------- |
| GET   | /api/me                    | Получить ID пользователя    |
| POST  | /api/accounts              | Создать счёт                |
| GET   | /api/accounts              | Список счетов               |
| POST  | /api/accounts/deposit      | Пополнение                  |
| POST  | /api/accounts/withdraw     | Списание                    |
| POST  | /api/transfer              | Перевод между счетами       |
| POST  | /api/cards                 | Генерация виртуальной карты |
| GET   | /api/cards                 | Получение списка карт       |
| POST  | /api/credits               | Оформление кредита          |
| GET   | /api/credits/{id}/schedule | График платежей по кредиту  |
| GET   | /api/accounts/{id}/predict | Прогноз баланса             |
| GET   | /api/analytics/credit-load | Кредитная нагрузка          |
| GET   | /api/test-email            | Тест email-уведомления      |
| GET   | /api/test-rate             | Ключевая ставка банка ЦБ    |

### HTML-страницы (для тестирования)

| Путь               | Назначение                 |
| ------------------ | -------------------------- |
| /register-form     | Регистрация                |
| /login-form        | Аутентификация             |
| /me-form           | Проверка токена            |
| /accounts-form     | Создание счёта             |
| /accounts-balance  | Пополнение / списание      |
| /transfer-form     | Перевод между счетами      |
| /cards-form        | Генерация карты            |
| /cards-view        | Просмотр карт              |
| /credits-form      | Оформление кредита         |
| /schedule-form     | Просмотр графика платежей  |
| /analytics-monthly | Аналитика доходов/расходов |
| /analytics-credit  | Анализ кредитной нагрузки  |
| /predict-form      | Прогноз баланса            |
| /test-email-form   | Тест отправки email        |
| /test-rate-form    | Запрос к банку ЦБ          |

## Используемые технологии

* Go 1.22+
* PostgreSQL 15+
* gorilla/mux (роутинг)
* JWT (аутентификация)
* bcrypt, PGP, HMAC (безопасность)
* gomail (email)
* etree (SOAP/XML)
* logrus (логирование)
