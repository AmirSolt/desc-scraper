package base

type Base struct {
	RootDir string
	*Env
	*Config
	*DB
}

func LoadBase() *Base {

	base := Base{
		RootDir: "./",
	}

	base.loadEnv()
	base.loadConfig()
	base.loadDB()

	return &base
}

func (base *Base) Kill() {
	base.killDB()
}
