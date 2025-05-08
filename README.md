# EasyREST Redis Cache Plugin

The **EasyREST Redis Cache Plugin** is an external plugin for [EasyREST](https://github.com/onegreyonewhite/easyrest) that enables EasyREST to use a Redis instance as a cache backend. This plugin implements the `easyrest.CachePlugin` interface.

**Key Features:**

- **Cache Operations:** Supports standard cache operations like SET (with Time-To-Live/TTL) and GET.
- **Redis Connection Pooling:** Utilizes the `go-redis/redis` library's built-in connection pooling for efficient Redis communication.
- **Flexible Configuration:** Supports standard Redis URI parameters and additional common connection pool settings via the connection URI.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Configuration Parameters](#configuration-parameters)
- [Redis Setup using Docker](#redis-setup-using-docker)
- [Environment Variables / Configuration for EasyREST](#environment-variables--configuration-for-easyrest)
- [Building the Plugin](#building-the-plugin)
- [Running EasyREST Server with the Plugin](#running-easyrest-server-with-the-plugin)
- [License](#license)

---

## Prerequisites

- [Docker](https://www.docker.com) installed on your machine.
- [Go 1.24](https://golang.org/dl/) or later (due to generics usage in dependencies).
- Basic knowledge of Redis and Docker.

---

## Configuration Parameters

The plugin configuration is managed via the Redis connection URI. It uses `redis.ParseURL` for standard parameters and supports additional query parameters for fine-tuning:

- **Standard URI Components:**
  - Scheme: `redis://`
  - User/Password: `user:password@`
  - Host and Port: `host:port`
  - Database Number: `/db_number` (e.g., `/0`, `/1`)
- **Query Parameters (Optional Overrides/Additions):**
  - `protocol`: Protocol version to use (2 or 3). Default is 3.
  - `client_name`: Name for the Redis client connection.
  - `max_retries`: Maximum number of retries before giving up (default: 3, -1 disables retries).
  - `min_retry_backoff`: Minimum backoff between each retry (e.g., `8ms`, -1 disables backoff).
  - `max_retry_backoff`: Maximum backoff between each retry (e.g., `512ms`, -1 disables backoff).
  - `dial_timeout`: Timeout for establishing new connections (e.g., `5s`, `100ms`).
  - `read_timeout`: Timeout for reading commands (e.g., `3s`).
  - `write_timeout`: Timeout for writing commands (e.g., `3s`).
  - `pool_fifo`: Use FIFO mode for connection pool (true/false).
  - `pool_size`: Maximum number of socket connections.
  - `pool_timeout`: Timeout for getting a connection from the pool.
  - `min_idle_conns`: Minimum number of idle connections.
  - `max_idle_conns`: Maximum number of idle connections.
  - `max_active_conns`: Maximum number of active connections.
  - `conn_max_idle_time` / `idle_timeout`: Maximum amount of time a connection may be idle (deprecated: `idle_timeout`).
  - `conn_max_lifetime` / `max_conn_age`: Maximum amount of time a connection may be reused (deprecated: `max_conn_age`).
  - `skip_verify`: If true, disables TLS certificate verification (for `rediss://`).

**Example URI with parameters:**

```
redis://user:password@myredishost:6379/1?dialTimeout=5s&readTimeout=2s&poolSize=100&minIdleConns=10
```

Refer to the [`go-redis` documentation](https://redis.uptrace.dev/guide/go-redis-option.html) for more details on connection options.

---

## Redis Setup using Docker

Run Redis in a Docker container. Open your terminal and execute:

```bash
docker run --name redis-easyrest -p 6379:6379 -d redis:7
```

This command starts a Redis 7 container:

- **Container Name:** `redis-easyrest`
- **Host Port:** `6379` (mapped to Redis's default port 6379 in the container)
- Uses the official Redis image.

For production, consider configuring persistence and security settings for Redis.

---

## Environment Variables / Configuration for EasyREST

Configure EasyREST to use this Redis instance via the Redis cache plugin. You can use environment variables or a configuration file.

**Using Environment Variables:**

```bash
# --- Cache Connection --- 
# The URI for the Redis cache. The plugin will be selected when the URI scheme is 'redis://'.
# Replace 'localhost' if Redis is running elsewhere.
export ER_CACHE_REDIS="redis://localhost:6379/0"

# Add other necessary EasyREST environment variables (e.g., for DB, auth)
# export ER_DB_MYSQL="mysql://..."
# export ER_TOKEN_SECRET="your-secret-key"
```

**Using a Configuration File (`config.yaml`):**

```yaml
plugins:
  redis_cache: # Or any name you choose for this cache instance
    uri: redis://localhost:6379/0?readTimeout=3s # URI for the Redis instance
    path: ./easyrest-plugin-redis # Path to the compiled Redis cache plugin binary

  # Define your primary database plugin (e.g., MySQL)
  # mysql:
  #   uri: mysql://user:pass@host:port/db
  #   cache_name: redis_cache
```

**Notes:**

- EasyREST identifies the plugin type based on the URI scheme (`redis://`).
- If using a config file, the `path` must point to the compiled plugin binary.
- Configuration file settings override environment variables for the same plugin instance name.
- The name used in the config file (e.g., `redis_cache`) or derived from the environment variable (e.g., `redis` from `ER_CACHE_REDIS`) identifies this specific plugin instance within EasyREST.

---

## Building the Plugin

Clone the repository for the EasyREST Redis Cache Plugin and build the plugin binary. In the repository root, run:

```bash
go build -o easyrest-plugin-redis redis_plugin.go
```

This produces the binary `easyrest-plugin-redis`.

---

## Running EasyREST Server with the Plugin

Download and install the pre-built binary for the EasyREST Server from the [EasyREST Releases](https://github.com/onegreyonewhite/easyrest/releases) page.

**Using a Configuration File (Recommended):**

1. **Create `config.yaml`:** Save your configuration (like the example above) to a file.
2. **Place Plugin Binary:** Ensure the compiled `easyrest-plugin-redis` binary is at the location specified by `path` in your config.
3. **Run Server:**
   ```bash
   ./easyrest-server --config config.yaml
   ```
   EasyREST will read the config, find the `plugins.redis_cache` section (or your chosen name), see the `redis://` URI, and load the plugin from the specified `path`.

**Using Environment Variables:**

1. **Set Environment Variables:** Define `ER_CACHE_REDIS` and any other required `ER_` variables.
2. **Place Plugin Binary:** The `easyrest-plugin-redis` binary *must* be in the same directory as `easyrest-server` or in your system `PATH`.
3. **Run Server:**
   ```bash
   ./easyrest-server
   ```
   EasyREST detects `ER_CACHE_REDIS`, looks for `easyrest-plugin-redis` in standard locations, and loads it if found.

Once running, EasyREST will use this plugin for its internal caching needs if caching is enabled for specific endpoints or globally.

---

## License

EasyREST Redis Cache Plugin is licensed under the Apache License 2.0.
See the file "LICENSE" for more information.

© 2025 Sergei Kliuikov
