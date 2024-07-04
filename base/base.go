package base

type Base struct {
	RootDir string
	Env     *Env
	Config  *Config
	DB      *DB
	MemQ    *MemcachedQueue
}

func LoadBase() *Base {

	base := Base{
		RootDir: "./",
	}

	base.loadEnv()
	base.loadConfig()
	base.loadDB()
	base.loadMemcached()

	return &base
}

func (base *Base) Kill() {
	base.killDB()
}
