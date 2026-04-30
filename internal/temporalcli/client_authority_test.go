package temporalcli

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
)

func TestApplyClientAuthorityFromConfig(t *testing.T) {
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
client_authority = "profile-authority"`))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	cctx := &CommandContext{
		RootCommand: &TemporalCommand{
			CommonOptions: cliext.CommonOptions{
				ConfigFile: f.Name(),
				Profile:    "foo",
			},
		},
	}
	var opts cliext.ClientOptions

	require.NoError(t, applyClientAuthorityFromConfig(cctx, &opts))
	require.Equal(t, "profile-authority", opts.ClientAuthority)
}

func TestApplyClientAuthorityFromConfig_ExplicitFlagWins(t *testing.T) {
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
client_authority = "profile-authority"`))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	cctx := &CommandContext{
		RootCommand: &TemporalCommand{
			CommonOptions: cliext.CommonOptions{
				ConfigFile: f.Name(),
				Profile:    "foo",
			},
		},
	}
	opts := cliext.ClientOptions{ClientAuthority: "flag-authority"}

	require.NoError(t, applyClientAuthorityFromConfig(cctx, &opts))
	require.Equal(t, "flag-authority", opts.ClientAuthority)
}

func TestApplyClientAuthorityFromConfig_ExplicitEmptyFlagWins(t *testing.T) {
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
client_authority = "profile-authority"`))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	cctx := &CommandContext{
		RootCommand: &TemporalCommand{
			CommonOptions: cliext.CommonOptions{
				ConfigFile: f.Name(),
				Profile:    "foo",
			},
		},
	}
	var opts cliext.ClientOptions
	opts.BuildFlags(pflag.NewFlagSet("test", pflag.ContinueOnError))
	require.NoError(t, opts.FlagSet.Set("client-authority", ""))

	require.NoError(t, applyClientAuthorityFromConfig(cctx, &opts))
	require.Empty(t, opts.ClientAuthority)
}
