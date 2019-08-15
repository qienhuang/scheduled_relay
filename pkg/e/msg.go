package e

var MsgFlags = map[int]string{
	SUCCESS:                        "ok",
	ERROR:                          "fail",
	INVALID_PARAMS:                 "INVALID_PARAMS",
	ERROR_EXIST_TAG:                "ERROR_EXIST_TAG",
	ERROR_NOT_EXIST_TAG:            "ERROR_NOT_EXIST_TAG",
	ERROR_NOT_EXIST_ARTICLE:        "ERROR_NOT_EXIST_ARTICLE",
	ERROR_AUTH_CHECK_TOKEN_FAIL:    "ERROR_AUTH_CHECK_TOKEN_FAIL",
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT: "ERROR_AUTH_CHECK_TOKEN_TIMEOUT",
	ERROR_AUTH_TOKEN:               "ERROR_AUTH_TOKEN",
	ERROR_AUTH:                     "ERROR_AUTH",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
