package shardserver

import (
	"chunkymonkey/object"
	. "chunkymonkey/types"
)

// localShardShardClient implements IShardShardClient for LocalShardManager.
type localShardShardClient struct {
	serverShard *ChunkShard
}

func newLocalShardShardClient(serverShard *ChunkShard) *localShardShardClient {
	return &localShardShardClient{
		serverShard: serverShard,
	}
}

func (client *localShardShardClient) Disconnect() {
}

func (client *localShardShardClient) ReqSetActiveBlocks(blocks []BlockXyz) {
	client.serverShard.Enqueue(func() {
		client.serverShard.reqSetBlocksActive(blocks)
	})
}

func (client *localShardShardClient) ReqTransferEntity(loc ChunkXz, entity object.INonPlayerEntity) {
	client.serverShard.Enqueue(func() {
		chunk := client.serverShard.Get(&loc)
		if chunk != nil {
			chunk.transferEntity(entity)
		}
	})
}