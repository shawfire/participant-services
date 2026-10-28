package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestJWKSCache(t *testing.T) {
	var requests atomic.Int32
	var version atomic.Int32
	version.Store(1)
	key1JSON := testJWKSJSON(t, "key-1")
	key2JSON := testJWKSJSON(t, "key-2")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if version.Load() == 1 {
				fmt.Fprint(w, key1JSON)
			} else {
				fmt.Fprint(w, key2JSON)
			}
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("initial fetch", func(t *testing.T) {
		keys, err := cache.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if keys.Keys[0].KeyID != "key-1" {
			t.Fatalf("unexpected key: %s", keys.Keys[0].KeyID)
		}
		if requests.Load() != 1 {
			t.Fatalf("expected 1 request, got %d", requests.Load())
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		_, err := cache.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if requests.Load() != 1 {
			t.Fatalf("cache miss: %d requests", requests.Load())
		}
	})

	t.Run("key rotation after expiry", func(t *testing.T) {
		version.Store(2)
		time.Sleep(60 * time.Millisecond)

		keys, err := cache.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if keys.Keys[0].KeyID != "key-2" {
			t.Fatalf("expected rotated key, got %s", keys.Keys[0].KeyID)
		}
		if requests.Load() != 2 {
			t.Fatalf("expected 2 requests, got %d", requests.Load())
		}
	})
}

func TestJWKSCacheConcurrentRefresh(t *testing.T) {
	var requests atomic.Int32
	keyJSON := testJWKSJSON(t, "key-1")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			fmt.Fprint(w, keyJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cache.Get(context.Background())
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	if requests.Load() != 1 {
		t.Fatalf(
			"expected one upstream request, got %d",
			requests.Load(),
		)
	}
}

func TestJWKSCacheFailsClosed(t *testing.T) {
	var fail atomic.Bool
	keyJSON := testJWKSJSON(t, "key-1")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if fail.Load() {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			fmt.Fprint(w, keyJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := cache.Get(context.Background()); err != nil {
		t.Fatal(err)
	}

	fail.Store(true)
	time.Sleep(30 * time.Millisecond)

	keys, err := cache.Get(context.Background())
	if err == nil {
		t.Fatal("expected expired cache refresh to fail")
	}
	if keys != nil {
		t.Fatal("expired keys must not be returned")
	}
}

func TestJWKSCacheInvalidConfiguration(t *testing.T) {
	if _, err := NewJWKSCache(nil, time.Minute); err == nil {
		t.Fatal("expected nil fetcher to be rejected")
	}

	fetcher, err := NewJWKSFetcher("https://example.com/keys", false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewJWKSCache(fetcher, 0); err == nil {
		t.Fatal("expected zero TTL to be rejected")
	}
}

func TestJWKSCacheMutationIsolation(t *testing.T) {
	keyJSON := testJWKSJSON(t, "key-1")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, keyJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	first, err := cache.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	first.Keys[0].KeyID = "attacker-controlled"
	first.Keys[0].Modulus = "invalid"
	first.Keys = append(first.Keys, JWK{
		KeyID: "injected",
	})

	second, err := cache.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(second.Keys) != 1 {
		t.Fatal("cached key collection was modified")
	}

	if second.Keys[0].KeyID != "key-1" {
		t.Fatal("cached key ID was modified")
	}

	if _, err := second.Keys[0].PublicKey(); err != nil {
		t.Fatalf("cached RSA key was modified: %v", err)
	}
}

func TestJWKSCacheRejectsInvalidRefresh(t *testing.T) {
	validJSON := testJWKSJSON(t, "key-1")

	var invalid atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if invalid.Load() {
				fmt.Fprint(w, `{"keys":[{"kid":"bad-key","kty":"RSA"}]}`)
				return
			}

			fmt.Fprint(w, validJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// First retrieval stores a valid RSA public key.
	keys, err := cache.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if keys.Keys[0].KeyID != "key-1" {
		t.Fatal("unexpected initial key")
	}

	// Simulate the identity provider publishing an invalid JWKS.
	invalid.Store(true)
	time.Sleep(30 * time.Millisecond)

	// Expired keys must not be returned when refresh fails.
	keys, err = cache.Get(ctx)
	if err == nil {
		t.Fatal("expected invalid JWKS refresh to fail")
	}

	if keys != nil {
		t.Fatal("expired keys were returned")
	}

	// Confirm the failed refresh did not overwrite stored keys.
	cache.mu.Lock()
	stored := cloneJWKS(cache.keys)
	cache.mu.Unlock()

	if stored == nil || len(stored.Keys) != 1 {
		t.Fatal("previously validated keys were lost")
	}

	if stored.Keys[0].KeyID != "key-1" {
		t.Fatal("invalid JWKS replaced the cached key")
	}

	if _, err := stored.Keys[0].PublicKey(); err != nil {
		t.Fatalf("previously cached key is invalid: %v", err)
	}
}

func TestJWKSCacheUnknownKeyRefreshRateLimit(t *testing.T) {
	keyJSON := testJWKSJSON(t, "key-1")

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, keyJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Initial JWKS retrieval.
	if _, err := cache.Get(ctx); err != nil {
		t.Fatal(err)
	}

	if requests.Load() != 1 {
		t.Fatalf("expected 1 request, got %d", requests.Load())
	}

	// First unknown key triggers an early refresh.
	if _, err := cache.RefreshForUnknownKey(
		ctx, "unknown-1",
	); err != nil {
		t.Fatalf("first early refresh failed: %v", err)
	}

	if requests.Load() != 2 {
		t.Fatalf("expected 2 requests, got %d", requests.Load())
	}

	// Repeated unknown keys must not trigger more requests.
	for i := 0; i < 20; i++ {
		_, err := cache.RefreshForUnknownKey(
			ctx,
			fmt.Sprintf("unknown-%d", i+2),
		)

		if err == nil {
			t.Fatal("expected early refresh to be rate limited")
		}
	}

	if requests.Load() != 2 {
		t.Fatalf(
			"rate limit failed: expected 2 requests, got %d",
			requests.Load(),
		)
	}

	// A known key should still be available from the cache.
	keys, err := cache.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := keys.FindPublicKey("key-1"); err != nil {
		t.Fatalf("known key unavailable: %v", err)
	}
}

func TestJWKSCacheConcurrentUnknownKeyRefresh(t *testing.T) {
	keyJSON := testJWKSJSON(t, "key-1")

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, keyJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Populate the cache with the original trusted key.
	if _, err := cache.Get(ctx); err != nil {
		t.Fatal(err)
	}

	if requests.Load() != 1 {
		t.Fatalf("expected initial fetch, got %d", requests.Load())
	}

	const callers = 20

	var wg sync.WaitGroup
	var successfulRefreshes atomic.Int32

	for i := 0; i < callers; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			_, err := cache.RefreshForUnknownKey(
				ctx,
				fmt.Sprintf("unknown-%d", index),
			)

			if err == nil {
				successfulRefreshes.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if successfulRefreshes.Load() != 1 {
		t.Fatalf(
			"expected one successful refresh, got %d",
			successfulRefreshes.Load(),
		)
	}

	if requests.Load() != 2 {
		t.Fatalf(
			"expected 2 total HTTP requests, got %d",
			requests.Load(),
		)
	}
}
