package engines

import (
	"fmt"
	"github.com/rhine-tech/scene"
)

const banner = `
===============================================================
                  ____                       
                 / ___|  ___ ___ _ __   ___   
                 \___ \ / __/ _ \ '_ \ / _ \ 
                  ___) | (_|  __/ | | |  __/  
                 |____/ \___\___|_| |_|\___|  
                                                        v%s
===============================================================
`

func getBanner() string {
	return fmt.Sprintf(banner, scene.Version)
}
