package parser_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"

	"github.com/crypto-com/chainindex/infrastructure/tendermint"
	"github.com/crypto-com/chainindex/usecase/parser"
	usecase_parser_test "github.com/crypto-com/chainindex/usecase/parser/test"
)

var _ = Describe("ParseMsgCommands", func() {
	Describe("MsgBeginRedelegate", func() {

		It("should parse Msg commands when there is staking.MsgEditValidator in the transaction", func() {
			txDecoder := parser.NewTxDecoder("basetrcro")
			block, _, _ := tendermint.ParseBlockResp(strings.NewReader(usecase_parser_test.TX_MSG_EDIT_VALIDATOR_BLOCK_RESP))
			blockResults, _ := tendermint.ParseBlockResultsResp(strings.NewReader(usecase_parser_test.TX_MSG_EDIT_VALIDATOR_BLOCK_RESULTS_RESP))

			cmds := parser.ParseMsgToCommands(
				txDecoder,
				block,
				blockResults,
			)
			fmt.Printf("%+v\n", cmds)
		})
	})
})
