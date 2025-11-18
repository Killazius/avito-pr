# README
## Структура проекта
Сервисы
**app** - основное приложение
- **Порт**: 8080
- Зависит от PostgreSQL и миграций
- Авторестарт при ошибках

**migrator** - запуск миграций БД
- Выполняется один раз при старте
- Зависит от **PostgreSQL**

**postgres** - база данных
- **Порт**: 5432
- Volume для данных
- Health check


### Multi-stage Dockerfile:
- **builder** - компиляция Go приложений
- **app** - финальный образ основного сервиса
- **migrator** - финальный образ для миграций

## Дополнительные задания выполнены:
- Реализовал эндпоинт статистики
`GET /stats`
Результатом является следующий JSON:
```json
{
    "total_teams": 2,
    "total_users": 5,
    "active_users": 4,
    "total_prs": 7,
    "open_prs": 7,
    "merged_prs": 0,
    "total_assignments": 9
}
```
- Настроил линтер golangci-lint

# Конфигурация
## .env.example
```text
CONFIG_PATH=config/config.yaml // Путь к файлу конфигурации

POSTGRES_HOST="postgres" // Имя хоста PostgreSQL
POSTGRES_PORT="5432" // Порт PostgreSQL
POSTGRES_USER="myuser" // Пользователь PostgreSQL
POSTGRES_PASSWORD="mypassword" // Пароль PostgreSQL
POSTGRES_DB="mydb" // Имя базы данных PostgreSQL
POSTGRES_SSL_MODE="disable" // Режим SSL для подключения к PostgreSQL

GIN_MODE=release // Режим работы Gin (debug/release/test)

GOOSE_DRIVER=postgres // Драйвер для Goose миграций
GOOSE_MIGRATION_DIR=./migrations // Путь к директории с миграциями
```

## Настройка окружения
1. `cp .env.example .env`
2. Заполните переменные окружения в файле .env при необходимости

# Запуск сервиса (после конфигурации)
```bash
make docker-build
make docker-up
```


# Makefile
- `docker-up`: Запускает Docker-контейнеры. (docker compose up)
- `docker-clean`: Останавливает и удаляет Docker-контейнеры и связанные с ними образы.
- `docker-build`: Собирает Docker-образы для всех сервисов, определенных в docker-compose.yaml.
- `test`: Запускает тесты Go с флагами для обнаружения гонок и сбора покрытия кода.
- `lint`: Запускает статический анализатор кода golangci-lint для всего проекта. 
- `deps`: Управляет зависимостями Go-модулей: загружает, проверяет и очищает их. 
- `build`: Компилирует приложение и сохраняет исполняемый файл в bin/app. 
- `clean`: Удаляет артефакты сборки, такие как директория bin/ и файл отчета о покрытии cover.out.
- `mock`: Генерирует моки для интерфейсов, определенных в проекте, используя mockery.