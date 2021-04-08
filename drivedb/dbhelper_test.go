package drivedb

import (
	dd "drurus/drivedb"
	"fmt"
	"testing"
)

func TestHost_ToStrings(t *testing.T) {
	host := dd.NewHost("testhost")
	st := host.ToStrings()
	fmt.Println(st)

}
