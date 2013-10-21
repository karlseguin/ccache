package ccache

type Configuration struct {
  size uint64
  buckets int
  itemsToPrune int
  promoteBuffer int
}

func Configure() *Configuration {
  return &Configuration{
    buckets: 64,
    itemsToPrune: 500,
    promoteBuffer: 1024,
    size: 500 * 1024 * 1024,
  }
}

func (c *Configuration) Size(bytes uint64) *Configuration {
  c.size = bytes
  return c
}

func (c *Configuration) Buckets(count int) *Configuration {
  c.buckets = count
  return c
}

func (c *Configuration) ItemsToPrune(count int) *Configuration {
  c.itemsToPrune = count
  return c
}

func (c *Configuration) PromoteBuffer(size int) *Configuration {
  c.promoteBuffer = size
  return c
}
