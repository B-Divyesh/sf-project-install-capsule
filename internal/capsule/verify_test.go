package capsule

import (
	"encoding/json"
	"testing"
)

func TestLastJSONLineIgnoresEngineProgress(t *testing.T) {
	output := []byte("Trying to pull image…\nCopying blob 1/2\n{\"host_secret_unreadable\":true}\n")
	var result VerifyResult
	if err := json.Unmarshal(lastJSONLine(output), &result); err != nil {
		t.Fatal(err)
	}
	if !result.HostSecretUnreadable {
		t.Fatal("verification result was not decoded")
	}
}
