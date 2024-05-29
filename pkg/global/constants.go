package global

const (
	WRAPPER_VERSION               = "0.0.20"
	STAGEENV                      = "stage"
	DEVENV                        = "dev"
	HTTPRESPONSETIMEOUT           = 60
	MAX_API_SERVER_START_ATTEMPTS = 10
	BadHandshake                  = "websocket: the client is not using the websocket protocol: "
)

var SOME_EXTERNAL_HOST map[string]string

func init() {
	SOME_EXTERNAL_HOST = map[string]string{
		"stage": "https://stage.someservice.com",
		"dev":   "https://dev.someservice.com",
		"prod":  "https://someservice.com",
	}

}
