package rabbitmq

func isNotValidTypeChan(typeChan uint64) bool {
	return typeChan == 0 || typeChan > 2
}

func isNotValidKey(key string) bool {
	return key == ""
}
