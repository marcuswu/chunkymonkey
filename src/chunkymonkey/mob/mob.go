package mob

import (
	"expvar"
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

var (
	expVarMobSpawnCount *expvar.Int
)

func init() {
	expVarMobSpawnCount = expvar.NewInt("mob-spawn-count")
}

// When using an object of type Mob or a sub-type, the caller must set an
// EntityId, most likely obtained from the EntityManager.
type Mob struct {
	EntityId
	mobType  EntityMobType
	look     LookDegrees
	metadata map[byte]byte
	// TODO: Change to an AABB object when we have that.
	physObj physics.PointObject
}

func (mob *Mob) Init(id EntityMobType, position *AbsXyz, velocity *AbsVelocity) {
	mob.physObj.Init(position, velocity)
	mob.mobType = id
	mob.look = LookDegrees{0, 0}
	mob.metadata = map[byte]byte{
		0:  byte(0),
		16: byte(0),
	}
}

func (mob *Mob) Position() *AbsXyz {
	return &mob.physObj.Position
}

func (mob *Mob) SetPosition(pos AbsXyz) {
	mob.physObj.Position = pos
}

func (mob *Mob) SetLook(look LookDegrees) {
	mob.look = look
}

func (mob *Mob) SetBurning(burn bool) {
	if burn {
		mob.metadata[0] |= 0x01
	} else {
		mob.metadata[0] ^= 0x01
	}
}

func (mob *Mob) Tick(blockQuery physics.BlockQueryFn) (leftBlock bool) {
	return mob.physObj.Tick(blockQuery)
}

func (mob *Mob) FormatMetadata() []proto.EntityMetadata {
	x := []proto.EntityMetadata{}
	for k, v := range mob.metadata {
		x = append(x, proto.EntityMetadata{0, k, v})
	}
	return x
}

func (mob *Mob) SendUpdate(writer io.Writer) (err os.Error) {
	if err = proto.WriteEntity(writer, mob.EntityId); err != nil {
		return
	}

	err = mob.physObj.SendUpdate(writer, mob.EntityId, &LookBytes{0, 0})

	return
}


func (mob *Mob) SendSpawn(writer io.Writer) (err os.Error) {
	err = proto.WriteEntitySpawn(
		writer,
		mob.EntityId,
		mob.mobType,
		&mob.physObj.LastSentPosition,
		mob.look.ToLookBytes(),
		mob.FormatMetadata())
	if err != nil {
		return
	}
	err = proto.WriteEntityVelocity(
		writer,
		mob.EntityId,
		&mob.physObj.LastSentVelocity)
	if err != nil {
		return
	}
	expVarMobSpawnCount.Add(1)

	return
}

// Evil mobs.

type Creeper struct {
	Mob
}

var (
	creeperNormal   = byte(0)
	creeperBlueAura = byte(1)
)

func NewCreeper(position *AbsXyz, velocity *AbsVelocity) (c *Creeper) {
	c = new(Creeper)
	c.Mob.Init(CreeperType.Id, position, velocity)
	c.Mob.metadata[17] = creeperNormal
	c.Mob.metadata[16] = byte(255)
	return c
}

func (c *Creeper) SetNormalStatus() {
	c.Mob.metadata[17] = creeperNormal
}

func (c *Creeper) CreeperSetBlueAura() {
	c.Mob.metadata[17] = creeperBlueAura
}

type Skeleton struct {
	Mob
}

func NewSkeleton(position *AbsXyz, velocity *AbsVelocity) (s *Skeleton) {
	s = new(Skeleton)
	s.Mob.Init(SkeletonType.Id, position, velocity)
	return
}


type Spider struct {
	Mob
}

func NewSpider(position *AbsXyz, velocity *AbsVelocity) (s *Spider) {
	s = new(Spider)
	s.Mob.Init(SpiderType.Id, position, velocity)
	return
}


// Passive mobs.

type Pig struct {
	Mob
}

func NewPig(position *AbsXyz, velocity *AbsVelocity) (p *Pig) {
	p = new(Pig)
	p.Mob.Init(PigType.Id, position, velocity)
	return
}

type Sheep struct {
	Mob
}

func NewSheep(position *AbsXyz, velocity *AbsVelocity) (s *Sheep) {
	s = new(Sheep)
	s.Mob.Init(SheepType.Id, position, velocity)
	return
}

type Cow struct {
	Mob
}

func NewCow(position *AbsXyz, velocity *AbsVelocity) (c *Cow) {
	c = new(Cow)
	c.Mob.Init(CowType.Id, position, velocity)
	return
}

type Hen struct {
	Mob
}

func NewHen(position *AbsXyz, velocity *AbsVelocity) (h *Hen) {
	h = new(Hen)
	h.Mob.Init(HenType.Id, position, velocity)
	return
}

type Squid struct {
	Mob
}

func NewSquid(position *AbsXyz, velocity *AbsVelocity) (s *Squid) {
	s = new(Squid)
	s.Mob.Init(SquidType.Id, position, velocity)
	return
}

type Wolf struct {
	Mob
}

func NewWolf(position *AbsXyz, velocity *AbsVelocity) (w *Wolf) {
	w = new(Wolf)
	w.Mob.Init(WolfType.Id, position, velocity)
	// TODO(nictuku): String with an optional owner's username.
	w.Mob.metadata[17] = 0
	w.Mob.metadata[16] = 0
	w.Mob.metadata[18] = 0
	return w
}
