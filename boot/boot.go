package boot

import (
	"github.com/ccb1900/gowebkit/db"
	"github.com/ccb1900/gowebkit/redisx"

	commonboot "github.com/ccb1900/gocommon/boot"
)

func Boot() {
	commonboot.Boot("test", "local")
	redisx.Init()
	db.Init()
}
