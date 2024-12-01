package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespace(t *testing.T) {
	oldPath := serviceAccountNamespaceFile

	t.Run("FallbackName", func(t *testing.T) {
		serviceAccountNamespaceFile = "not-a-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.Nil(err)
		assert.Equal(namespaceFleetlock, ns)
	})
	t.Run("ReadFromFile", func(t *testing.T) {
		serviceAccountNamespaceFile = "ns-from-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		err := os.WriteFile(serviceAccountNamespaceFile, []byte("success"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			err = os.Remove(serviceAccountNamespaceFile)
			t.Log(serviceAccountNamespaceFile)
			if err != nil {
				t.Log(err)
			}
		})

		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.Nil(err)
		assert.Equal("success", ns)
	})
	t.Run("FileEmpty", func(t *testing.T) {
		serviceAccountNamespaceFile = "ns-from-empty-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		err := os.WriteFile(serviceAccountNamespaceFile, []byte(""), 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			err = os.Remove(serviceAccountNamespaceFile)
			t.Log(serviceAccountNamespaceFile)
			if err != nil {
				t.Log(err)
			}
		})

		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.NotNil(err)
		assert.Equal("", ns)
	})
}

func TestPointer(t *testing.T) {
	s := "test"
	p := Pointer(s)
	assert.Equal(t, &s, p, "Should return pointer to variable with the same value")
}
