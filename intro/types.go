package intro

import (
	sky "github.com/skycoin/cxo/skyobject"
)

type Content struct {
	Thread sky.Refs `skyobject:"schema=intro.Vote"`
	Post   sky.Refs `skyobject:"schema=intro.Vote"`
}

// TODO: what it means?

// type Votes struct {
// 	Votes sky.Refs `skyobject:"schema="intro.Vote"`
// }

type Vote struct {
	Up    bool
	Index uint32
}
