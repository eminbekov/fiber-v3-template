package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// CacheClear clears Redis cache by prefix or flushes all keys.
func CacheClear(ctx context.Context, dependencies *Dependencies, arguments []string) error {
	flagSet := flag.NewFlagSet("cache-clear", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	prefixValue := flagSet.String("prefix", "", "optional cache prefix to clear")
	if parseError := flagSet.Parse(arguments); parseError != nil {
		return fmt.Errorf("cache-clear parse flags: %w", parseError)
	}

	normalizedPrefix := strings.TrimSpace(*prefixValue)
	if normalizedPrefix != "" {
		if deleteByPrefixError := dependencies.Cache.DeleteByPrefix(ctx, normalizedPrefix); deleteByPrefixError != nil {
			return fmt.Errorf("cache-clear delete by prefix: %w", deleteByPrefixError)
		}
		fmt.Fprintf(os.Stderr, "cleared cache with prefix %q\n", normalizedPrefix)
		return nil
	}

	if flushError := dependencies.RedisClient.FlushDB(ctx).Err(); flushError != nil {
		return fmt.Errorf("cache-clear flush db: %w", flushError)
	}
	fmt.Fprintln(os.Stderr, "cleared entire redis cache")
	return nil
}
