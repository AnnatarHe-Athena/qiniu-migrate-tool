package config

type cfg struct {
	IsDEV bool
	Host,
	Username,
	Pwd,
	Dbname,
	AccessKey,
	SecretKey,
	Bucket string
}

type Cell struct {
	ID  int
	Src string
}

func GetConfig() cfg {
	return cfg{
		Host:     "localhost",
		Username: "postgres",
		Pwd:      "admin",
		Dbname:   "postgres",

		AccessKey: "pkZuicLquRswTTjygYXFgpUfVzpvIwa3-i5cHfe-",
		SecretKey: "BG6vZZFYimQqHgZmvmuzXSb8KK_5A5b0HZf0l9iB",
		Bucket:    "iamhele-com",
		IsDEV:     true,
	}
}
