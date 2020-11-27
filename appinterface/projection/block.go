package projection

import (
	"fmt"

	entity_projection "github.com/crypto-com/chainindex/entity/projection"

	"github.com/crypto-com/chainindex/appinterface/projection/rdbbase"
	"github.com/crypto-com/chainindex/appinterface/projection/view"
	"github.com/crypto-com/chainindex/appinterface/rdb"
	event_entity "github.com/crypto-com/chainindex/entity/event"
	applogger "github.com/crypto-com/chainindex/internal/logger"
	event_usecase "github.com/crypto-com/chainindex/usecase/event"
)

var _ entity_projection.Projection = &Block{}

// TODO: Listen to council node related events and project council node
type Block struct {
	*rdbbase.RDbBase

	rdbConn rdb.Conn
	logger  applogger.Logger
}

func NewBlock(logger applogger.Logger, rdbConn rdb.Conn) *Block {
	return &Block{
		rdbbase.NewRDbBase(rdbConn.ToHandle(), "Block"),

		rdbConn,
		logger,
	}
}

func (_ *Block) GetEventsToListen() []string {
	return []string{event_usecase.BLOCK_CREATED}
}

func (projection *Block) OnInit() error {
	return nil
}

func (projection *Block) HandleEvents(height int64, events []event_entity.Event) error {
	var err error

	var rdbTx rdb.Tx
	rdbTx, err = projection.rdbConn.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	rdbTxHandle := rdbTx.ToHandle()
	blocksView := view.NewBlocks(rdbTxHandle)

	for _, event := range events {
		if blockCreatedEvent, ok := event.(*event_usecase.BlockCreated); ok {
			if err = projection.handleBlockCreatedEvent(blocksView, blockCreatedEvent); err != nil {
				_ = rdbTx.Rollback()
				return fmt.Errorf("error handling BlockCreatedEvent: %v", err)
			}
		} else {
			_ = rdbTx.Rollback()
			return fmt.Errorf("received unexpected event %sV%d(%s)", event.Name(), event.Version(), event.UUID())
		}
	}
	if err = projection.UpdateLastHandledEventHeight(rdbTxHandle, height); err != nil {
		_ = rdbTx.Rollback()
		return fmt.Errorf("error updating last handled event height: %v", err)
	}

	if err = rdbTx.Commit(); err != nil {
		_ = rdbTx.Rollback()
		return fmt.Errorf("error committing changes: %v", err)
	}
	return nil
}

func (projection *Block) handleBlockCreatedEvent(blocksView *view.Blocks, event *event_usecase.BlockCreated) error {
	committedCouncilNodes := make([]view.BlockCommittedCouncilNode, 0)
	for _, signature := range event.Block.Signatures {
		committedCouncilNodes = append(committedCouncilNodes, view.BlockCommittedCouncilNode{
			Address:    signature.ValidatorAddress,
			Time:       signature.Timestamp,
			Signature:  signature.Signature,
			IsProposer: event.Block.ProposerAddress == signature.ValidatorAddress,
		})
	}

	if err := blocksView.Insert(&view.Block{
		Height:                event.Block.Height,
		Hash:                  event.Block.Hash,
		Time:                  event.Block.Time,
		AppHash:               event.Block.AppHash,
		TransactionCount:      len(event.Block.Txs),
		CommittedCouncilNodes: committedCouncilNodes,
	}); err != nil {
		return err
	}

	return nil
}