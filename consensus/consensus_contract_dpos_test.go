package consensus

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/consensus/core"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/pool/lock"
)

func TestContractDposCs_ElectionIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	mock_chain := NewMockChain(ctrl)

	mock_chain.EXPECT().GetGenesisSnapshotBlock().Return(&ledger.SnapshotBlock{
		Timestamp: &simpleGenesis,
	})
	db := NewDb(t, UnitTestDir)
	defer ClearDb(t, UnitTestDir)
	mock_chain.EXPECT().NewDb(gomock.Any()).Return(db, nil)

	group := types.ConsensusGroupInfo{
		Gid:                    types.DELEGATE_GID,
		NodeCount:              2,
		Interval:               1,
		PerCount:               3,
		RandCount:              1,
		RandRank:               100,
		Repeat:                 1,
		CountingTokenId:        ledger.ViteTokenId,
		RegisterConditionId:    0,
		RegisterConditionParam: nil,
		VoteConditionId:        0,
		VoteConditionParam:     nil,
		Owner:                  types.Address{},
		StakeAmount:            nil,
		ExpirationHeight:       0,
	}

	info := core.NewGroupInfo(simpleGenesis, group)

	b1 := GenSnapshotBlock(1, "3fc5224e59433bff4f48c83c0eb4edea0e4c42ea697e04cdec717d03e50d5200", types.Hash{}, simpleGenesis)

	rw := newChainRw(mock_chain, log15.New(), &lock.EasyImpl{})

	cs := newContractDposCs(info, rw, log15.New())

	voteTime := cs.GenProofTime(0)
	mock_chain.EXPECT().GetSnapshotHeaderBeforeTime(gomock.Eq(&voteTime)).Return(b1, nil)
	registers := []*types.Registration{{
		Name:                  "s1",
		BlockProducingAddress: common.MockAddress(0),
		StakeAddress:          common.MockAddress(0),
		Amount:                nil,
		ExpirationHeight:      0,
		RewardTime:            0,
		RevokeTime:            0,
		HisAddrList:           nil,
	}, {
		Name:                  "s2",
		BlockProducingAddress: common.MockAddress(1),
		StakeAddress:          common.MockAddress(1),
		Amount:                nil,
		ExpirationHeight:      0,
		RewardTime:            0,
		RevokeTime:            0,
		HisAddrList:           nil,
	}}
	votes := []*types.VoteInfo{
		{
			VoteAddr: common.MockAddress(11),
			SbpName:  "s1",
		},
		{
			VoteAddr: common.MockAddress(12),
			SbpName:  "s1",
		}, {
			VoteAddr: common.MockAddress(21),
			SbpName:  "s2",
		}}

	S1balances := make(map[types.Address]*big.Int)
	S1balances[common.MockAddress(11)] = big.NewInt(11)
	S1balances[common.MockAddress(12)] = big.NewInt(12)
	S2balances := make(map[types.Address]*big.Int)
	S2balances[common.MockAddress(21)] = big.NewInt(21)

	mock_chain.EXPECT().GetRegisterList(b1.Hash, types.DELEGATE_GID).Return(registers, nil)
	mock_chain.EXPECT().GetVoteList(b1.Hash, types.DELEGATE_GID).Return(votes, nil)
	mock_chain.EXPECT().GetConfirmedBalanceList([]types.Address{common.MockAddress(11), common.MockAddress(12)}, ledger.ViteTokenId, b1.Hash).Return(S1balances, nil)
	mock_chain.EXPECT().GetConfirmedBalanceList([]types.Address{common.MockAddress(21)}, ledger.ViteTokenId, b1.Hash).Return(S2balances, nil)
	mock_chain.EXPECT().GetRandomSeed(b1.Hash, 25).Return(uint64(105))

	result, err := cs.ElectionIndex(0)
	assert.NoError(t, err)

	assert.NotNil(t, result)

	assert.Equal(t, simpleGenesis, result.STime)
	assert.Equal(t, simpleGenesis.Add(time.Duration(info.PlanInterval)*time.Second), result.ETime)
	assert.Equal(t, uint64(0), result.Index)
	assert.Equal(t, 6, len(result.Plans))
	for k, v := range result.Plans {
		assert.Equal(t, simpleGenesis.Add(time.Duration(int64(k)*info.Interval)*time.Second), v.STime)
		assert.Equal(t, v.STime.Add(time.Second), v.ETime)
		assert.Equal(t, common.MockAddress(k/int(info.PerCount)%int(info.NodeCount)), v.Member, fmt.Sprintf("%d", k))
	}
}
