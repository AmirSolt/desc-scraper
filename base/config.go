package base

type Config struct {
	MaxQueueSize int
}

func (b *Base) loadConfig() {
	config := Config{
		MaxQueueSize: 10_000_000,
	}

	b.Config = &config
}
