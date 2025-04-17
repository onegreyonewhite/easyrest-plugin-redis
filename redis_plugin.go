package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"strings"
	"time"

	hplugin "github.com/hashicorp/go-plugin"
	easyrest "github.com/onegreyonewhite/easyrest/plugin"
	"github.com/redis/go-redis/v9"
)

// Version can be set during build time
var Version = "v0.1.0"

// redisCachePlugin implements the CachePlugin interface using Redis.
type redisCachePlugin struct {
	client *redis.Client
}

// InitConnection establishes a connection to the Redis server based on the URI.
// Supports URI parameters like: db, password, dialTimeout, readTimeout, writeTimeout, poolSize, minIdleConns, idleTimeout.
func (p *redisCachePlugin) InitConnection(uri string) error {
	if !strings.HasPrefix(uri, "redis://") {
		return errors.New("invalid Redis URI: must start with redis://")
	}

	// Use redis.ParseURL which handles standard Redis URIs correctly.
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URI: %w", err)
	}

	// Extract custom parameters (if any) that redis.ParseURL might not handle directly
	// or allow overriding standard ones if needed.
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("failed to re-parse URI for custom params: %w", err) // Should not happen if redis.ParseURL succeeded
	}
	queryParams := parsedURI.Query()

	// Example: Allow overriding timeouts via query params (redis.ParseURL might already set some)
	if val := queryParams.Get("dialTimeout"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			opts.DialTimeout = d
		} else {
			return fmt.Errorf("invalid dialTimeout value: %s", val)
		}
	}
	if val := queryParams.Get("readTimeout"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			opts.ReadTimeout = d
		} else {
			return fmt.Errorf("invalid readTimeout value: %s", val)
		}
	}
	if val := queryParams.Get("writeTimeout"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			opts.WriteTimeout = d
		} else {
			return fmt.Errorf("invalid writeTimeout value: %s", val)
		}
	}
	/* Pool options removed for v9
	if val := queryParams.Get("poolSize"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			opts.PoolSize = i
		} else {
			return fmt.Errorf("invalid poolSize value: %s", val)
		}
	}
	if val := queryParams.Get("minIdleConns"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 0 {
			opts.MinIdleConns = i
		} else {
			return fmt.Errorf("invalid minIdleConns value: %s", val)
		}
	}
	if val := queryParams.Get("idleTimeout"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			opts.IdleTimeout = d // This field is removed in v9
		} else {
			return fmt.Errorf("invalid idleTimeout value: %s", val)
		}
	}
	*/
	// Note: The 'db' parameter is typically handled by the path in the URI for redis.ParseURL (e.g., redis://host:port/1)
	// If provided as a query parameter (e.g., ?db=1), redis.ParseURL should handle it. If not, add logic here.

	p.client = redis.NewClient(opts)

	// Ping the server to ensure connection is working.
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout+opts.ReadTimeout+1*time.Second) // Use a reasonable combined timeout
	defer cancel()

	status := p.client.Ping(ctx)
	if err := status.Err(); err != nil {
		p.client.Close() // Close the client if ping fails
		p.client = nil
		return fmt.Errorf("failed to ping Redis server: %w", err)
	}

	return nil
}

// Set stores a key-value pair with a TTL in the Redis cache.
func (p *redisCachePlugin) Set(key string, value string, ttl time.Duration) error {
	if p.client == nil {
		return errors.New("redis client not initialized")
	}
	ctx := context.Background() // Use background context for cache operations
	err := p.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache entry in Redis for key '%s': %w", key, err)
	}
	return nil
}

// Get retrieves a value from the Redis cache.
// It returns redis.Nil if the key is not found.
func (p *redisCachePlugin) Get(key string) (string, error) {
	if p.client == nil {
		return "", errors.New("redis client not initialized")
	}
	ctx := context.Background()
	val, err := p.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Return redis.Nil directly as go-redis v9 recommends.
			// EasyREST core or the caller can handle this specific error.
			return "", redis.Nil // Changed from errors.New("cache miss")
		}
		return "", fmt.Errorf("failed to get cache entry from Redis for key '%s': %w", key, err)
	}
	return val, nil
}

// Close cleans up the Redis client connection.
func (p *redisCachePlugin) Close() error {
	if p.client != nil {
		err := p.client.Close()
		p.client = nil // Ensure client is marked as nil after closing
		if err != nil {
			return fmt.Errorf("error closing Redis client: %w", err)
		}
	}
	return nil
}

func main() {
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(Version)
		return
	}

	cacheImpl := &redisCachePlugin{}

	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: easyrest.Handshake,
		Plugins: map[string]hplugin.Plugin{
			// Only register the cache plugin
			"cache": &easyrest.CachePluginPlugin{Impl: cacheImpl},
		},
	})
	defer cacheImpl.Close()
}
