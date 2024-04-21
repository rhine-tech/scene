package database

import "github.com/rhine-tech/scene"

type Database interface {
	DatabaseName() scene.ImplName
}
