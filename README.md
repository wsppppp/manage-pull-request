1) Вопрос использования фреймворков... выбор между gin и chi...


Для запуска проекта на локальной машине вам понадобятся **Docker** и **Docker Compose**.

1.  **Клонируйте репозиторий:**
    ```bash
    git clone https://github.com/wsppppp/manage-pull-request.git
    cd manage-pull-request
    ```

2.  **Создайте файл конфигурации:**
    Скопируйте файл-шаблон `.env.example` в новый файл `.env`. В этом файле хранятся параметры для подключения к базе данных. Для локального запуска значения по умолчанию подходят идеально.
    ```bash
    # Для Linux/macOS
    cp .env.example .env

    # Для Windows (в cmd)
    copy .env.example .env
    ```

3.  **Запустите проект с помощью Docker Compose:**
    Эта команда соберет Go-приложение, поднимет контейнер с PostgreSQL, применит миграции базы данных и запустит сервис.

    ```bash
    docker-compose up --build -d
    ```
    *   Флаг `--build` пересобирает образ приложения, если в коде были изменения.
    *   Флаг `-d` запускает контейнеры в фоновом режиме.

    Сервис будет доступен по адресу `http://localhost:8080`.

4.  **Остановка проекта:**
    ```bash
    docker-compose down
    ```
    Чтобы остановить контейнеры и удалить том с данными PostgreSQL (полная очистка), используйте:
    ```bash
    docker-compose down -v
    ```

---


# Шаг 1: Чистый старт (рекомендуется)
docker-compose down -v && docker-compose up --build -d

# --- Пауза в несколько секунд ---

# Шаг 2: Создание команды
curl --location 'http://localhost:8080/team/add' \
--header 'Content-Type: application/json' \
--data '{
"name": "sre-team",
"members": [
{"user_id": "u1", "username": "Alice", "is_active": true},
{"user_id": "u2", "username": "Bob", "is_active": true},
{"user_id": "u3", "username": "Charlie", "is_active": true},
{"user_id": "u4", "username": "Diana", "is_active": true}
]
}'

# --- Пауза ---

# Шаг 3: Создание Pull Request'а
curl --location 'http://localhost:8080/pullRequest/create' \
--header 'Content-Type: application/json' \
--data '{
"pull_request_id": "PR-42",
"pull_request_name": "Refactor logging module",
"author_id": "u1"
}'

# --- Пауза ---

# Шаг 4: Проверка ревью для одного из назначенных (предположим, это u2)
curl --location 'http://localhost:8080/users/getReview?user_id=u2'

# --- Пауза ---

# Шаг 5: Переназначение ревью с u2
curl --location 'http://localhost:8080/pullRequest/reassign' \
--header 'Content-Type: application/json' \
--data '{
"pull_request_id": "PR-42",
"old_user_id": "u2"
}'

# --- Пауза ---

# Шаг 6: Проверка, что у u4 появилось ревью
curl --location 'http://localhost:8080/users/getReview?user_id=u4'

# --- Пауза ---

# Шаг 7: Мерж PR
curl --location 'http://localhost:8080/pullRequest/merge' \
--header 'Content-Type: application/json' \
--data '{
"pull_request_id": "PR-42"
}'