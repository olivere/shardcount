package v6

import (
	"context"
	"log"
	"os"

	std "github.com/olivere/elastic/v7"

	"github.com/olivere/shardcount/elastic"
)

type Client struct {
	c *std.Client
}

func NewClient(cfg elastic.Config) (*Client, error) {
	options := []std.ClientOptionFunc{
		std.SetURL(cfg.URL),
		std.SetSniff(cfg.Sniff),
		std.SetHealthcheck(cfg.Healthcheck),
	}
	if cfg.Trace {
		options = append(options, std.SetTraceLog(log.New(os.Stdout, "", 0)))
	}
	es, err := std.NewClient(options...)
	if err != nil {
		return nil, err
	}
	c := &Client{
		c: es,
	}
	return c, nil
}

func (c *Client) ListShards(ctx context.Context, req elastic.ListShardsRequest) (elastic.ListShardsResponse, error) {
	shards, err := c.c.CatShards().Index(req.Index).Do(ctx)
	if err != nil {
		return nil, err
	}
	var resp elastic.ListShardsResponse
	for _, shard := range shards {
		resp = append(resp, elastic.ListShardsIndexAndShard{
			Index: shard.Index,
			Shard: shard.Shard,
			Docs:  shard.Docs,
		})
	}
	return resp, nil
}
