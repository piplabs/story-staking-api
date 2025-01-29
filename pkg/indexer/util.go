package indexer

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
)

func attrArray2Map(attrs []abcitypes.EventAttribute) map[string]string {
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	return attrMap
}
