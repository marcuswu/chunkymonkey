// Map chunks

package chunkymonkey

import (
    "bytes"
    "io"
    "os"
    "log"
    "path"
    "nbt/nbt"
)

// A chunk is slice of the world map
type Chunk struct {
    mgr        *ChunkManager
    XZ         ChunkXZ
    Blocks     []byte
    BlockData  []byte
    BlockLight []byte
    SkyLight   []byte
    HeightMap  []byte
}

func blockIndex(subLoc *SubChunkXYZ) (index int32, shift byte, err bool) {
    if subLoc.x < 0 || subLoc.y < 0 || subLoc.z < 0 || subLoc.x >= ChunkSizeX || subLoc.y >= ChunkSizeY || subLoc.z >= ChunkSizeZ {
        err = true
        index = 0
    } else {
        err = false

        index = int32(subLoc.y + (subLoc.z * ChunkSizeY) + (subLoc.x * ChunkSizeY * ChunkSizeZ))

        if index%2 == 0 {
            // Low nibble
            shift = 0
        } else {
            // High nibble
            shift = 4
        }
    }
    return
}

// Sets a block and its data. Returns true if the block was not changed.
func (chunk *Chunk) SetBlock(subLoc *SubChunkXYZ, blockType BlockID, blockMetadata byte) (err bool) {
    index, shift, err := blockIndex(subLoc)
    if err {
        return
    }

    chunk.Blocks[index] = byte(blockType)

    mask := byte(0x0f) << shift
    twoBlockData := chunk.BlockData[index/2]
    twoBlockData = ((blockMetadata << shift) & mask) | (twoBlockData & ^mask)
    chunk.BlockData[index/2] = twoBlockData

    // Tell players that the block was destroyed
    packet := &bytes.Buffer{}
    WriteBlockChange(packet, chunk.XZ.ToBlockXY(subLoc), BlockIDAir, 0)
    chunk.mgr.game.MulticastChunkPacket(packet.Bytes(), chunk.XZ)

    return
}

// Returns information about the block at the given location. err is true if
// subLoc is outside of the chunk.
func (chunk *Chunk) GetBlock(subLoc *SubChunkXYZ) (blockType BlockID, err bool) {
    index, _, err := blockIndex(subLoc)
    if err {
        return
    }

    blockType = BlockID(chunk.Blocks[index])
    return
}

// Send chunk data down network connection
func (chunk *Chunk) SendChunkData(writer io.Writer) (err os.Error) {
    return WriteMapChunk(writer, &chunk.XZ, chunk.Blocks, chunk.BlockData, chunk.BlockLight, chunk.SkyLight)
}

// Load a chunk from its NBT representation
func loadChunk(reader io.Reader) (chunk *Chunk, err os.Error) {
    level, err := nbt.Read(reader)
    if err != nil {
        return
    }

    chunk = &Chunk{
        XZ: ChunkXZ{
            x:  ChunkCoord(level.Lookup("/Level/xPos").(*nbt.Int).Value),
            z:  ChunkCoord(level.Lookup("/Level/zPos").(*nbt.Int).Value),
        },
        Blocks:     level.Lookup("/Level/Blocks").(*nbt.ByteArray).Value,
        BlockData:  level.Lookup("/Level/Data").(*nbt.ByteArray).Value,
        SkyLight:   level.Lookup("/Level/SkyLight").(*nbt.ByteArray).Value,
        BlockLight: level.Lookup("/Level/BlockLight").(*nbt.ByteArray).Value,
        HeightMap:  level.Lookup("/Level/HeightMap").(*nbt.ByteArray).Value,
    }
    return
}

// ChunkManager contains all chunks and can look them up
type ChunkManager struct {
    game      *Game
    worldPath string
    chunks    map[uint64]*Chunk
}

func NewChunkManager(worldPath string) *ChunkManager {
    return &ChunkManager{
        worldPath: worldPath,
        chunks:    make(map[uint64]*Chunk),
    }
}

func base36Encode(n int32) (s string) {
    alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
    negative := false

    if n < 0 {
        n = -n
        negative = true
    }
    if n == 0 {
        return "0"
    }

    for n != 0 {
        i := n % int32(len(alphabet))
        n /= int32(len(alphabet))
        s = string(alphabet[i:i+1]) + s
    }
    if negative {
        s = "-" + s
    }
    return
}

func (mgr *ChunkManager) chunkPath(loc ChunkXZ) string {
    return path.Join(mgr.worldPath, base36Encode(int32(loc.x&63)), base36Encode(int32(loc.z&63)),
        "c."+base36Encode(int32(loc.x))+"."+base36Encode(int32(loc.z))+".dat")
}

// Get a chunk at given coordinates
func (mgr *ChunkManager) Get(loc ChunkXZ) (chunk *Chunk) {
    // FIXME this function looks subject to race conditions with itself
    key := uint64(loc.x)<<32 | uint64(uint32(loc.z))
    chunk, ok := mgr.chunks[key]
    if ok {
        return
    }

    file, err := os.Open(mgr.chunkPath(loc), os.O_RDONLY, 0)
    if err != nil {
        log.Exit("ChunkManager.Get: ", err.String())
    }

    chunk, err = loadChunk(file)
    chunk.mgr = mgr
    file.Close()
    if err != nil {
        log.Exit("ChunkManager.loadChunk: ", err.String())
    }

    mgr.chunks[key] = chunk
    return
}

// Return a channel to iterate over all chunks within a chunk's radius
func (mgr *ChunkManager) ChunksInRadius(loc *ChunkXZ) (c chan *Chunk) {
    c = make(chan *Chunk)
    go func() {
        curChunkXZ := ChunkXZ{0, 0}
        for z := loc.z - ChunkRadius; z <= loc.z+ChunkRadius; z++ {
            for x := loc.x - ChunkRadius; x <= loc.x+ChunkRadius; x++ {
                curChunkXZ.x, curChunkXZ.z = x, z
                c <- mgr.Get(curChunkXZ)
            }
        }
        close(c)
    }()
    return
}

// Return a channel to iterate over all chunks within a player's radius
func (mgr *ChunkManager) ChunksInPlayerRadius(player *Player) chan *Chunk {
    playerChunkXZ := player.position.ToChunkXZ()
    return mgr.ChunksInRadius(&playerChunkXZ)
}
