//go:build windows
// +build windows

package dummy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	dummyServiceLog = "dummyservice.log"
)

// doDummyServiceWork demos Windows-specific service work
func (svc *Service) doDummyServiceWork() {
	dummyData := &dummyData{
		OwningSvc:      svc.name,
		DummyData:      "windows dummy data" + randomString(8),
		CollectionTime: time.Now().Format(time.RFC3339),
	}
	dummyDataBytes, err := json.Marshal(&dummyData)
	if err != nil {
		return
	}
	absPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return
	}
	dir := filepath.Dir(absPath)
	dummySvcLogFilePath := filepath.Join(dir, dummyServiceLog)
	os.WriteFile(dummySvcLogFilePath, dummyDataBytes, 0666)
}
