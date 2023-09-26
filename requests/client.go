package requests

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/KakashiHatake324/tlsclient/v2"
	tls "github.com/KakashiHatake324/tlsclient/v2/utls"
)

// create a client with a given proxy
func CreateRequestClient(proxy string) *http.Client {
	jar, _ := cookiejar.New(nil)
	settings := tlsclient.CustomizedSettings{

		MaxHeaderListSize: 262144,

		// Set as true to include enable push in frames
		ServerPushSet: true,

		// Set as true to set enable push value to 1
		// or set as false to set enable push value to 0
		// if ServerPushSet is not as true this will not get sent.
		ServerPushEnable: false,

		Priority: true,

		// Set value from 1 to 256
		PriorityWeight:       256,
		InitialWindowSize:    6291456,
		MaxConcurrentStreams: 1000,
		HeaderTableSize:      65536,
		WindowSizeIncrement:  15663105,
	}

	var client = new(http.Client)
	*client, _ = tlsclient.NewClient(tls.HelloChrome_100, jar, false, time.Duration(10*time.Second), settings, "", "", proxy)
	return client
}
