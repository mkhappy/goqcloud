package qcloud

type Config struct {
	SecretId      string `json:"secretId"`
	SecretKey     string `json:"secretKey"`
	DefaultRegion string `json:"defaultRegion"`
	RequestMethod string `json:"requestMethod"`
}
