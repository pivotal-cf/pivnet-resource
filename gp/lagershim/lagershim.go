package lagershim

import (
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-golang/lager"
)

type LagerShim interface {
	Debug(action string, data ...logger.Data)
	Info(action string, data ...logger.Data)
}

type lagerShim struct {
	l lager.Logger
}

func NewLagerShim(l lager.Logger) LagerShim {
	return &lagerShim{
		l: l,
	}
}

func (l lagerShim) Debug(action string, data ...logger.Data) {
	allLagerData := mapData(data...)
	l.l.Debug(action, allLagerData...)
}

func mapData(data ...logger.Data) []lager.Data {
	allLagerData := make([]lager.Data, len(data))

	for i, d := range data {
		allLagerData[i] = lager.Data(d)
	}

	return allLagerData
}

func (l lagerShim) Info(action string, data ...logger.Data) {
	allLagerData := mapData(data...)
	l.l.Info(action, allLagerData...)
}
