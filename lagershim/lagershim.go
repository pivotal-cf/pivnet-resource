package lagershim

import (
	"fmt"

	gpl "github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

type LagerShim interface {
	Debug(action string, data ...gpl.Data)
	Info(action string, data ...gpl.Data)
}

type lagerShim struct {
	l logger.Logger
}

func NewLagerShim(l logger.Logger) LagerShim {
	return &lagerShim{
		l: l,
	}
}

func (l lagerShim) Debug(action string, data ...gpl.Data) {
	allLagerData := mapData(data...)
	l.l.Debugf(action, allLagerData)
}

func mapData(data ...gpl.Data) string {
	return fmt.Sprintf("%+v", data)
}

func (l lagerShim) Info(action string, data ...gpl.Data) {
	allLagerData := mapData(data...)
	l.l.Debugf(action, allLagerData)
}
