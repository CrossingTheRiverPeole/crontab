package common

//定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

/**
构建返回结果
 */
func BuildResponse(errno int, msg string, data interface{}) (resp map[string]interface{}) {
	m := make(map[string]interface{})
	//错误编码
	m["errno"] = errno
	// 错误信息
	m["msg"] = msg
	// 数据
	m["data"] = data
	return m
}
