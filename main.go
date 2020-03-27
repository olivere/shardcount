package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Masterminds/semver"

	"github.com/olivere/shardcount/elastic"
	v5 "github.com/olivere/shardcount/elastic/v5"
	v6 "github.com/olivere/shardcount/elastic/v6"
	v7 "github.com/olivere/shardcount/elastic/v7"
)

func main() {
	if err := runMain(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runMain() error {
	var (
		url         = flag.String("url", "http://localhost:9200", "URL to Elasticsearch")
		index       = flag.String("index", "", "Index pattern")
		sniff       = flag.Bool("sniff", false, "Enable or disable sniff")
		healthcheck = flag.Bool("healthcheck", false, "Enable or disable health checks")
		trace       = flag.Bool("trace", false, "Enable or disable trace logging to standard output")
	)
	flag.Parse()

	cfg := elastic.Config{
		URL:         *url,
		Sniff:       *sniff,
		Healthcheck: *healthcheck,
		Trace:       *trace,
	}
	client, err := newClient(cfg)
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	// List all shards
	shards, err := client.ListShards(context.Background(), elastic.ListShardsRequest{
		Index: *index,
	})
	if err != nil {
		return fmt.Errorf("unable to list shards: %w", err)
	}

	// Group by index and shards
	docsPerShard := make(map[string]int64) // -> index+shard -> docs
	for _, shard := range shards {
		key := fmt.Sprintf("%s#%d", shard.Index, shard.Shard)
		if n, ok := docsPerShard[key]; !ok {
			docsPerShard[key] = shard.Docs
		} else if n != shard.Docs {
			fmt.Fprintf(os.Stdout, "%s already registered with %d documents, but shard %d has %d\n", shard.Index, shard.Docs, shard.Shard, n)
		}
	}

	return nil
}

func newClient(cfg elastic.Config) (elastic.Client, error) {
	v, major, _, _, err := elasticsearchVersion(cfg)
	if err != nil {
		return nil, err
	}
	switch major {
	default:
		return nil, fmt.Errorf("no Elasticsearch client for version %s", v)
	case 5:
		return v5.NewClient(cfg)
	case 6:
		return v6.NewClient(cfg)
	case 7:
		return v7.NewClient(cfg)
	}
}

func elasticsearchVersion(cfg elastic.Config) (string, int64, int64, int64, error) {
	type infoType struct {
		Name    string `json:"name"`
		Version struct {
			Number string `json:"number"` // e.g. "6.2.4"
		} `json:"version"`
	}
	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		return "", 0, 0, 0, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, 0, 0, err
	}
	defer res.Body.Close()
	var info infoType
	if err = json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", 0, 0, 0, err
	}
	v, err := semver.NewVersion(info.Version.Number)
	if err != nil {
		return info.Version.Number, 0, 0, 0, err
	}
	return info.Version.Number, v.Major(), v.Minor(), v.Patch(), nil
}
