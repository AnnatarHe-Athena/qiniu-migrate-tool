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
	Md5 string
}

func GetConfig() cfg {
	return cfg{

		Host:      "115.159.89.54",
		Username:  "postgres",
		Pwd:       "d8fdd2eb46a84a7c13656f012d118760ff43aef1e580e8577c7bcd157abac033",
		Dbname:    "postgres",
		AccessKey: "pkZuicLquRswTTjygYXFgpUfVzpvIwa3-i5cHfe-",
		SecretKey: "BG6vZZFYimQqHgZmvmuzXSb8KK_5A5b0HZf0l9iB",

		// Host:      "localhost",
		// Username:  "postgres",
		// Pwd:       "admin",
		// Dbname:    "postgres",
		// AccessKey: "",
		// SecretKey: "",

		Bucket: "iamhele-com",
		IsDEV:  true,
	}
}
