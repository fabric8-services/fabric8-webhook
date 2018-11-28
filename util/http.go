package util

import (
	"net/http"
	"time"
)

// NetClient defines the default HTTP client
var NetClient = &http.Client{
	Timeout: time.Second * 10,
}
