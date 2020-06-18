package main

// Config is ...
type Config struct {
	DeployDir        string `env:"BITRISE_DEPLOY_DIR"`
	AABPath          string `env:"aab_path,required"`
	KeystorePath     string `env:"keystore_path"`
	KeystotePassword string `env:"keystore_password"`
	KeyAlias         string `env:"key_alias"`
	KeyPassword      string `env:"key_password"`
}
