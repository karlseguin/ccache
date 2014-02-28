package ccache

type Configuration struct {
	size           uint64
	buckets        int
	itemsToPrune   int
	deleteBuffer   int
	promoteBuffer  int
	getsPerPromote int32
	tracking       bool
}

func Configure() *Configuration {
	return &Configuration{
		buckets:        64,
		itemsToPrune:   500,
		deleteBuffer:   1024,
		getsPerPromote: 10,
		promoteBuffer:  1024,
		size:           500 * 1024 * 1024,
		tracking:       false,
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

func (c *Configuration) DeleteBuffer(size int) *Configuration {
	c.deleteBuffer = size
	return c
}

func (c *Configuration) GetsPerPromote(count int) *Configuration {
	c.getsPerPromote = int32(count)
	return c
}

func (c *Configuration) Track() *Configuration {
	c.tracking = true
	return c
}
