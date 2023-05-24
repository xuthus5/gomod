package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalUrlGetTag(t *testing.T) {
	output, err := lsRemote.setUrl("https://gitter.top/coco/bootstrap").tagOrCommitID()
	assert.NoError(t, err)
	t.Log(output)

	output, err = lsRemote.setUrl("https://github.com/spf13/cobra").tagOrCommitID()
	assert.NoError(t, err)
	t.Log(output)
}
