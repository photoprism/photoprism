package classify

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var onceDs sync.Once
var testInstanceDs *DeepStack

// NewTest returns a new DeepStack test instance.
func DeepStackNewTest(t *testing.T) *DeepStack {
	onceDs.Do(func() {
		testInstanceDs = DeepStackNew("http://host.docker.internal:5000/", false)
		if err := testInstanceDs.DeepStackInit(); err != nil {
			t.Fatal(err)
		}

	})
	return testInstanceDs
}

func TestDeepStack_LabelsFromFile(t *testing.T) {
	t.Run("cat_black.jpg", func(t *testing.T) {
		deepStack := DeepStackNewTest(t)

		result, err := deepStack.DeepStackFile(examplesPath + "/cat_black.jpg")

		assert.Nil(t, err)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 2, len(result))

		t.Log(result)

		assert.Equal(t, "cat", result[0].Name)

		assert.Equal(t, 40, result[0].Uncertainty)
	})
	t.Run("not existing file", func(t *testing.T) {
		deepStack := DeepStackNewTest(t)

		result, err := deepStack.DeepStackFile(examplesPath + "/notexisting.jpg")
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
	t.Run("disabled true", func(t *testing.T) {
		deepStack := DeepStackNew("", true)

		result, err := deepStack.DeepStackFile(examplesPath + "/cat_black.jpg")
		assert.Nil(t, err)

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 0, len(result))

		t.Log(result)
	})
}
