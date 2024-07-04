package base

type Config struct {
	MaxQueueSize int
}

func (b *Base) loadConfig() {
	config := Config{
		MaxQueueSize: 1000,
	}

	b.Config = &config
}
