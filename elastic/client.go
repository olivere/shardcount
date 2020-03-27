package elastic

import "context"

type Client interface {
	ListShards(context.Context, ListShardsRequest) (ListShardsResponse, error)
}

type ListShardsRequest struct {
	Index string
}

type ListShardsResponse []ListShardsIndexAndShard

type ListShardsIndexAndShard struct {
	Index string
	Shard int
	Docs  int64
}
