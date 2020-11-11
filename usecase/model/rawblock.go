package model

import (
	"github.com/crypto-com/chainindex/internal/utctime"
)

// RawBlock defines the structure for Tendermint /block API response JSON
type RawBlock struct {
	BlockID struct {
		Hash  string `json:"hash" fake:"{blockheight}"`
		Parts struct {
			Total int    `json:"total"`
			Hash  string `json:"hash"`
		} `json:"parts"`
	} `json:"block_id"`
	Block struct {
		Header struct {
			Version struct {
				Block string `json:"block"`
			} `json:"version"`
			ChainID     string          `json:"chain_id"`
			Height      string          `json:"height"`
			Time        utctime.UTCTime `json:"time"`
			LastBlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total int    `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"last_block_id"`
			LastCommitHash     string `json:"last_commit_hash"`
			DataHash           string `json:"data_hash"`
			ValidatorsHash     string `json:"validators_hash"`
			NextValidatorsHash string `json:"next_validators_hash"`
			ConsensusHash      string `json:"consensus_hash"`
			AppHash            string `json:"app_hash"`
			LastResultsHash    string `json:"last_results_hash"`
			EvidenceHash       string `json:"evidence_hash"`
			ProposerAddress    string `json:"proposer_address"`
		} `json:"header"`
		Data struct {
			Txs []string `json:"txs"`
		} `json:"data"`
		Evidence struct {
			Evidence []struct {
				Type  string `json:"type"`
				Value struct {
					PubKey struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"PubKey"`
					VoteA struct {
						Type    int    `json:"type"`
						Height  string `json:"height"`
						Round   int    `json:"round"`
						BlockID struct {
							Hash  string `json:"hash"`
							Parts struct {
								Total int    `json:"total"`
								Hash  string `json:"hash"`
							} `json:"parts"`
						} `json:"block_id"`
						Timestamp        utctime.UTCTime `json:"timestamp"`
						ValidatorAddress string          `json:"validator_address"`
						ValidatorIndex   string          `json:"validator_index"`
						Signature        string          `json:"signature"`
					} `json:"VoteA"`
					VoteB struct {
						Type    int    `json:"type"`
						Height  string `json:"height"`
						Round   int    `json:"round"`
						BlockID struct {
							Hash  string `json:"hash"`
							Parts struct {
								Total int    `json:"total"`
								Hash  string `json:"hash"`
							} `json:"parts"`
						} `json:"block_id"`
						Timestamp        utctime.UTCTime `json:"timestamp"`
						ValidatorAddress string          `json:"validator_address"`
						ValidatorIndex   string          `json:"validator_index"`
						Signature        string          `json:"signature"`
					} `json:"VoteB"`
				} `json:"value"`
			} `json:"evidence"`
		} `json:"evidence"`
		LastCommit struct {
			Height  string `json:"height"`
			Round   int    `json:"round"`
			BlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total int    `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Signatures []RawBlockSignature `json:"signatures"`
		} `json:"last_commit"`
	} `json:"block"`
}

// RawBlockSignature defines the structure for signatures in /block API
type RawBlockSignature struct {
	BlockIDFlag      int             `json:"block_id_flag"`
	ValidatorAddress string          `json:"validator_address"`
	Timestamp        utctime.UTCTime `json:"timestamp"`
	MaybeSignature   *string         `json:"signature"`
}
