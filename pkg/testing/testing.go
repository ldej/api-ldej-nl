package testing

import (
	"os"
	"path"
	"runtime"
)

// Import in tests to change working directory to root
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
