package registry

import "reflect"

// DisposeAll dispose all disposable objects
// dangerous function, only use when program ends
func DisposeAll() {
	for _, disposable := range Disposable.AcquireAll() {
		err := disposable.Dispose()
		if err != nil {
			Logger.Warnf("dispose %v error: %v", reflect.TypeOf(disposable), err)
		}
	}

}

func SetupAll() {
	for _, setupable := range Setupable.AcquireAll() {
		err := setupable.Setup()
		if err != nil {
			Logger.Warnf("setup %v error: %v", reflect.TypeOf(setupable), err)
		}
	}
}
