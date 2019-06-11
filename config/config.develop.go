// +build !release

package config

type Cell struct {
	ID  int
	Src string
	Md5 string
}

const (
	Host            = "localhost"
	Username        = "postgres"
	Pwd             = "admin"
	Dbname          = "postgres"
	AccessKey       = ""
	SecretKey       = ""
	Bucket          = "iamhele-com"
	IsDEV           = true
	TencentAIAppID  = ""
	TencentAIAppKey = ""
	// TencentAIAppID = "2117314147"
	// TencentAIAppKey = "r9CTbgv065oi8QJl"
)
