package log

import (
	"bytes"
	"fmt"
	"github.com/grrrben/distsys/registry"
	stlog "log"
	"net/http"
)

func SetClientLogger(serviceUrl string, clientService registry.ServiceName) {
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	stlog.SetFlags(0)
	stlog.SetOutput(&clientLogger{url: serviceUrl})
}

type clientLogger struct {
	url string
}

func (cl clientLogger) Write(p []byte) (int, error) {
	res, err := http.Post(cl.url+"/log", "text/plain", bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to sent log data to %v/log, server returned with code %d", cl.url, res.StatusCode)
	}

	return len(p), nil
}
