package services

import (
	"cardGame/model/packet"
	"context"
)

type (
	SummonService struct {
	}
)

func NewSummonService() *SummonService {
	return &SummonService{}
}

func (s *SummonService) SummonCard(ctx context.Context, c *packet.SummonCard) {

}
