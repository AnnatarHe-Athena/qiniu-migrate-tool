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
		Host:      "localhost",
		Username:  "postgres",
		Pwd:       "admin",
		Dbname:    "postgres",
		AccessKey: "",
		SecretKey: "",

		Bucket: "iamhele-com",
		IsDEV:  true,
	}
}
