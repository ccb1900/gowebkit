package boot

import (
	"gowebkit/db"
	"gowebkit/redisx"

	commonboot "github.com/ccb1900/gocommon/boot"
)

func Boot() {
	commonboot.Boot("test", "local")
	redisx.Init()
	db.Init()
}
