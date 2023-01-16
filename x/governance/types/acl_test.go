package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACLGetSetOwner(t *testing.T) {
	acl := ACL(make([]ACLPair, 0))
	a := getRandomValidatorAddress()
	acl.SetOwner("governance/acl", a)
	assert.Equal(t, acl.GetOwner("governance/acl").String(), a.String())
}

func TestValidateACL(t *testing.T) {
	acl := createTestACL()
	adjMap := createTestAdjacencyMap()
	assert.Nil(t, acl.Validate(adjMap))
	acl.SetOwner("governance/acl2", getRandomValidatorAddress())
	assert.NotNil(t, acl.Validate(adjMap))
}
