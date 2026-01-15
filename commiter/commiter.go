package commiter

import (
	"fmt"

	"github.com/peiblow/vvm/vm"
)

type MockCommitter struct{}

func (m *MockCommitter) Commit(journal []vm.JournalEvent) error {
	fmt.Println("📦 Committing journal:")
	for _, e := range journal {
		fmt.Printf(" - %s %+v\n", e.Type, e.Payload)
	}
	return nil
}
