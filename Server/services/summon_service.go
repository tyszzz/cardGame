package services

import (
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
)

type (
	SummonService struct {
	}
)

func NewSummonService() *SummonService {
	return &SummonService{}
}

func (s *SummonService) SummonCard(ctx context.Context, data *packet.SummonCard) (*packet.SummonCardResult, error) {
	utils.Log().Infof("SummonCard called, UID: %s, Times: %d, SummonType: %v", data.UID, data.Times, data.SummonType)
	return &packet.SummonCardResult{
		UID:      data.UID,
		CardId:   "123",
		CardName: "123",
	}, nil
}
