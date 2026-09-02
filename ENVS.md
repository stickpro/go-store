# Environments

| **Name**                       | **Required** | **Secret** | **Default value**        | **Usage**                              | **Example**              |
|--------------------------------|--------------|------------|--------------------------|----------------------------------------|--------------------------|
| `APP_PROFILE`                  |              |            | `dev`                    |                                        |                          |
| `HTTP_HOST`                    |              |            | `localhost`              |                                        |                          |
| `HTTP_PORT`                    |              |            | `80`                     |                                        |                          |
| `HTTP_CONNECT_TIMEOUT`         |              |            | `5s`                     |                                        |                          |
| `HTTP_READ_TIMEOUT`            |              |            | `10s`                    |                                        |                          |
| `HTTP_WRITE_TIMEOUT`           |              |            | `10s`                    |                                        |                          |
| `HTTP_MAX_HEADER_MEGABYTES`    |              |            | `1`                      |                                        |                          |
| `HTTP_CORS_ENABLED`            |              |            | `true`                   | allows to disable cors                 | `true / false`           |
| `HTTP_CORS_ALLOWED_ORIGINS`    |              |            |                          |                                        |                          |
| `HTTP_MAX_BODY_LIMIT`          |              |            | `100`                    | maximum body size in mb, default 100MB | `100`                    |
| `POSTGRES_ADDR`                | ✅            |            |                          | storage address in format ip:port      |                          |
| `POSTGRES_DB_NAME`             | ✅            |            |                          | storage user                           |                          |
| `POSTGRES_USER`                | ✅            | ✅          |                          | storage user password                  |                          |
| `POSTGRES_PASSWORD`            | ✅            | ✅          |                          | storage db name                        |                          |
| `POSTGRES_CONN_MAX_LIFETIME`   |              |            | `180`                    |                                        |                          |
| `POSTGRES_MAX_OPEN_CONNS`      |              |            | `100`                    |                                        |                          |
| `POSTGRES_MAX_IDLE_CONNS`      |              |            | `100`                    |                                        |                          |
| `POSTGRES_MIN_OPEN_CONNS`      |              |            | `6`                      |                                        |                          |
| `REDIS_ADDR`                   | ✅            |            |                          | storage address in format ip:port      |                          |
| `REDIS_USER`                   |              | ✅          |                          | storage user password                  |                          |
| `REDIS_PASSWORD`               |              | ✅          |                          | storage db name                        |                          |
| `REDIS_DB_INDEX`               |              |            |                          | index of database                      |                          |
| `LOG_FORMAT`                   |              |            | `json`                   | allows to set custom formatting        | `json`                   |
| `LOG_LEVEL`                    |              |            | `info`                   | allows to set custom logger level      | `info`                   |
| `LOG_CONSOLE_COLORED`          |              |            | `false`                  | allows to set colored console output   | `false`                  |
| `LOG_TRACE`                    |              |            | `fatal`                  | allows to set custom trace level       | `fatal`                  |
| `LOG_WITH_CALLER`              |              |            | `false`                  | allows to show caller                  | `false`                  |
| `LOG_WITH_STACK_TRACE`         |              |            | `false`                  | allows to show stack trace             | `false`                  |
| `KEY_VALUE_ENGINE`             | ✅            |            | `redis`                  |                                        | `redis / in_memory`      |
| `FILE_STORAGE_TYPE`            |              |            | `local`                  |                                        | `local / s3`             |
| `FILE_STORAGE_PATH`            |              |            | `storage`                |                                        |                          |
| `FILE_STORAGE_BUCKET`          |              |            |                          |                                        |                          |
| `SEARCH_ENGINE_TYPE`           |              |            | `meili_search`           |                                        | `meili_search / elastic` |
| `SEARCH_ENGINE_HOST`           |              |            | `http://localhost:7700`  |                                        |                          |
| `SEARCH_ENGINE_API_KEY`        |              |            |                          |                                        |                          |
| `KAFKA_BROKERS`                |              |            |                          |                                        |                          |
| `KAFKA_TOPICS_PRODUCTS`        |              |            | `store.products`         |                                        |                          |
| `KAFKA_TOPICS_VARIANTS`        |              |            | `store.product-variants` |                                        |                          |
| `KAFKA_PRODUCER_REQUIRED_ACKS` |              |            | `all`                    |                                        |                          |
| `KAFKA_CONSUMER_GROUP_ID`      |              |            | `go-store`               |                                        |                          |
| `WORKERS_IMAGE_SYNC`           |              |            | `3`                      |                                        |                          |
| `IMAGES_JPEG_QUALITY`          |              |            | `82`                     |                                        |                          |
| `IMAGES_WEBP_QUALITY`          |              |            | `80`                     |                                        |                          |
| `IMAGES_MAX_SOURCE_PIXELS`     |              |            | `40000000`               |                                        |                          |
| `IMAGES_PRESETS`               |              |            |                          |                                        |                          |
