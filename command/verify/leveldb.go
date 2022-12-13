package verify

import (
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
)

func newBadgerBuilder(log hclog.Logger, path string) kvdb.BadgerBuilder {
	badgerBuilder := kvdb.NewBadgerBuilder(
		log,
		path,
	)

	return badgerBuilder
}
