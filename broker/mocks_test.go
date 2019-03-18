package broker_test

import (
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/mock"
)

type (
	MockMemberlist struct {
		mock.Mock
	}
)

func (m *MockMemberlist) NumMembers() int {
	return m.Called().Int(0)
}

func (m *MockMemberlist) LocalNode() *memberlist.Node {
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0).(*memberlist.Node)
	}

	return nil
}

func (m *MockMemberlist) Members() []*memberlist.Node {
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0).([]*memberlist.Node)
	}

	return nil
}
