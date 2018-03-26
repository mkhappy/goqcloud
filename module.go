package qcloud

func get_know_module_host(name string) string {
	switch name {
	case "wenzhi":
		return "wenzhi.api.qcloud.com"
	}
	return ""
}
