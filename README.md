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
## config/config.yaml
```yaml
server: # Настройки сервера
  host: 0.0.0.0 # Хост
  port: 8080 # Порт
  timeout: # Таймауты
    server: 30s # Общий таймаут
    read: 15s # Таймаут чтения
    write: 10s # Таймаут записи
    idle: 5s # Таймаут простоя
postgres: # Доп. Настройки PostgreSQL
  max_open_conns: 25 # Максимальное количество открытых соединений
  max_idle_conns: 5 # Максимальное количество неактивных соединений
  conn_max_lifetime: 1h # Максимальное время жизни соединения
  conn_max_idle_time: 25m # Максимальное время простоя соединения
  timeout: 5s # Таймаут подключения
```
## config/logger.json
```json
{
  "level": "debug",  // Уровень логирования (при GIN_MODE=release меняется на info)
  "encoding": "json",
  "outputPaths": ["stdout"],
  "errorOutputPaths": ["stderr"],
  "encoderConfig": {
    "timeKey": "timestamp",
    "timeEncoder": "rfc3339",
    "messageKey": "message",
    "levelKey": "level",
    "levelEncoder": "lowercase",
    "callerKey": "caller",
    "callerEncoder": "short"
  }
}
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