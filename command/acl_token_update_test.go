package command

import (
	"os"
	"testing"

	"github.com/hashicorp/nomad/acl"
	"github.com/hashicorp/nomad/command/agent"
	"github.com/hashicorp/nomad/nomad/mock"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func TestACLTokenUpdateCommand(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()
	config := func(c *agent.Config) {
		c.ACL.Enabled = true
	}

	srv, _, url := testServer(t, true, config)
	defer srv.Shutdown()

	// Bootstrap an initial ACL token
	token := srv.Token
	assert.NotNil(token, "failed to bootstrap ACL token")

	ui := new(cli.MockUi)
	cmd := &ACLTokenUpdateCommand{Meta: Meta{Ui: ui, flagAddress: url}}
	state := srv.Agent.Server().State()

	// Create a valid token
	mockToken := mock.ACLToken()
	mockToken.Policies = []string{acl.PolicyWrite}
	mockToken.SetHash()
	assert.Nil(state.UpsertACLTokens(1000, []*structs.ACLToken{mockToken}))

	// Request to update a new token without providing a valid management token
	invalidToken := mock.ACLToken()
	os.Setenv("NOMAD_TOKEN", invalidToken.SecretID)
	code := cmd.Run([]string{"-address=" + url, "-name=bar", mockToken.AccessorID})
	//code := cmd.Run([]string{"-address=" + url, "-policy=foo", "-type=client"})
	assert.Equal(1, code)

	// Request to update a new token with a valid management token
	os.Setenv("NOMAD_TOKEN", token.SecretID)
	code = cmd.Run([]string{"-address=" + url, "-name=bar", mockToken.AccessorID})
	assert.Equal(0, code)

	// Check the output
	out := ui.OutputWriter.String()
	assert.Contains(out, mockToken.AccessorID)
	assert.Contains(out, "bar")
}
