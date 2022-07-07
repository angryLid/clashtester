package tester

import (
	"testing"
)

func TestMockMain(t *testing.T) {

	for range [5]int{} {
		tester := NewTester(UseMock())
		tester.SetItem(ItemUserInfo | ItemLatency | ItemDownload | ItemUpload)
		result, _ := tester.Do()
		t.Log(result)
	}
}
