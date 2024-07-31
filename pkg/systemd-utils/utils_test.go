package systemdutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZincatiMachineID(t *testing.T) {
	tMatrix := []struct {
		MachineID string
		Result    string
	}{
		{"dfd7882acda64c34aca76193c46f5d4e", "35ba2101ae3f4d45b96e9c51f461bbff"},
		{"37974b3f7dc54f949209b4fd5b3c5704", "f59c4fa7d80e406f83993d64abf922e3"},
		{"4473d601f9234ff2a84617c3eaeeea35", "7742030900754495bfeb49c7d1f4d653"},
		{"10c30367c3cb44eaaab432e552061395", "72fe241381444dee85d824b8ec33a70f"},
		{"b0236093845b4344930e60dff4d355b0", "3a711518ca8542cb9249bf4fd859a5e1"},
		{"384afb4f366a4afbbe07c6fd0b9b222a", "473ca41a469040bfb5968791e929d581"},
		{"6a1ed0358a1b4e8eb16d2872a91358ce", "e9aa1d2389a541ba87046c5a2445b48d"},
	}

	for i, tCase := range tMatrix {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			assert := assert.New(t)

			res, err := ZincatiMachineID(tCase.MachineID)

			assert.Nil(err)
			assert.Equal(tCase.Result, res)
		})
	}
}

func TestAppSpecificID(t *testing.T) {
	assert := assert.New(t)

	res1, err1 := AppSpecificID("Not a hex string", "123456")
	assert.Empty(res1)
	assert.NotNil(err1)

	res2, err2 := AppSpecificID("123456", "Not a hex string")

	assert.Empty(res2)
	assert.NotNil(err2)
}
