package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialRepo(t *testing.T) {
	for in, want := range map[string]string{
		"protocol=https\nhost=github.com\npath=gfx-labs/jobshop.git\n\n": "gfx-labs/jobshop",
		"protocol=https\nhost=github.com\npath=acme/site\n":              "acme/site",
		"protocol=https\nhost=github.com\n":                              "",
		"protocol=https\nhost=gitlab.com\npath=a/b\n":                    "",
	} {
		got, ok := credentialRepo(parseCredential(strings.NewReader(in)))
		assert.Equal(t, want != "", ok, in)
		assert.Equal(t, want, got, in)
	}
}
