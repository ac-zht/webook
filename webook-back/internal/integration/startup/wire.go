package startup

import "github.com/google/wire"

var thirdProvider = wire.NewSet(InitRedis)
