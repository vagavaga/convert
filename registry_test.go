package convert_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vagavaga/convert"
)

func TestRegister(t *testing.T) {
	var s string
	var i int
	var sliceS []string
	var sliceI []int
	var err error

	r := convert.Registry{}
	err = r.Register("test")
	assert.Error(t, err)

	r.Register(func(s string, i *int) error {
		c, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*i = c
		return nil
	})
	r.Register(func(i int, s *string) {
		*s = strconv.Itoa(i)
		// panic("boom")
	})

	err = r.Convert("10", &s)
	if assert.NoError(t, err) {
		assert.Equal(t, "10", s)
	}
	err = r.Convert("10", &i)
	if assert.NoError(t, err) {
		assert.Equal(t, 10, i)
	}
	err = r.Convert(9, &i)
	if assert.NoError(t, err) {
		assert.Equal(t, 9, i)
	}
	err = r.Convert(9, &s)
	if assert.NoError(t, err) {
		assert.Equal(t, "9", s)
	}

	err = r.Convert([]string{"1", "2", "3"}, &sliceS)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"1", "2", "3"}, sliceS)
	}

	err = r.Convert([]string{"4", "5", "6"}, &sliceI)
	if assert.NoError(t, err) {
		assert.Equal(t, []int{4, 5, 6}, sliceI)
	}

	err = r.Convert([]int{2, 4, 6}, &sliceI)
	if assert.NoError(t, err) {
		assert.Equal(t, []int{2, 4, 6}, sliceI)
	}

	err = r.Convert([]int{1, 3, 5}, &sliceS)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"1", "3", "5"}, sliceS)
	}

	// fmt.Println("r", &r)
}
