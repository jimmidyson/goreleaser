package cmd

import (
	"path/filepath"
	"testing"

	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestRelease(t *testing.T) {
	setup(t)
	cmd := newReleaseCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2", "--deprecated"})
	require.NoError(t, cmd.cmd.Execute())
}

func TestReleaseAutoSnapshot(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		setup(t)
		cmd := newReleaseCmd()
		cmd.cmd.SetArgs([]string{"--auto-snapshot", "--skip-publish"})
		require.NoError(t, cmd.cmd.Execute())
		require.FileExists(t, "dist/fake_0.0.2_checksums.txt", "should not have run with --snapshot")
	})

	t.Run("dirty", func(t *testing.T) {
		setup(t)
		createFile(t, "foo", "force dirty tree")
		cmd := newReleaseCmd()
		cmd.cmd.SetArgs([]string{"--auto-snapshot", "--skip-publish"})
		require.NoError(t, cmd.cmd.Execute())
		matches, err := filepath.Glob("./dist/fake_v0.0.2-SNAPSHOT-*_checksums.txt")
		require.NoError(t, err)
		require.Len(t, matches, 1, "should have implied --snapshot")
	})
}

func TestReleaseInvalidConfig(t *testing.T) {
	setup(t)
	createFile(t, "goreleaser.yml", "foo: bar")
	cmd := newReleaseCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2", "--deprecated"})
	require.EqualError(t, cmd.cmd.Execute(), "yaml: unmarshal errors:\n  line 1: field foo not found in type config.Project")
}

func TestReleaseBrokenProject(t *testing.T) {
	setup(t)
	createFile(t, "main.go", "not a valid go file")
	cmd := newReleaseCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2"})
	require.EqualError(t, cmd.cmd.Execute(), "failed to parse dir: .: main.go:1:1: expected 'package', found not")
}

func TestReleaseFlags(t *testing.T) {
	setup := func(opts releaseOpts) *context.Context {
		return setupReleaseContext(context.New(config.Project{}), opts)
	}

	t.Run("snapshot", func(t *testing.T) {
		ctx := setup(releaseOpts{
			snapshot: true,
		})
		require.True(t, ctx.Snapshot)
		require.True(t, ctx.SkipPublish)
		require.True(t, ctx.SkipValidate)
		require.True(t, ctx.SkipAnnounce)
	})

	t.Run("skips", func(t *testing.T) {
		ctx := setup(releaseOpts{
			skipPublish:  true,
			skipSign:     true,
			skipValidate: true,
		})
		require.True(t, ctx.SkipSign)
		require.True(t, ctx.SkipPublish)
		require.True(t, ctx.SkipValidate)
		require.True(t, ctx.SkipAnnounce)
	})

	t.Run("parallelism", func(t *testing.T) {
		require.Equal(t, 1, setup(releaseOpts{
			parallelism: 1,
		}).Parallelism)
	})

	t.Run("notes", func(t *testing.T) {
		notes := "foo.md"
		header := "header.md"
		footer := "footer.md"
		ctx := setup(releaseOpts{
			releaseNotesFile:  notes,
			releaseHeaderFile: header,
			releaseFooterFile: footer,
		})
		require.Equal(t, notes, ctx.ReleaseNotesFile)
		require.Equal(t, header, ctx.ReleaseHeaderFile)
		require.Equal(t, footer, ctx.ReleaseFooterFile)
	})

	t.Run("templated notes", func(t *testing.T) {
		notes := "foo.md"
		header := "header.md"
		footer := "footer.md"
		ctx := setup(releaseOpts{
			releaseNotesTmpl:  notes,
			releaseHeaderTmpl: header,
			releaseFooterTmpl: footer,
		})
		require.Equal(t, notes, ctx.ReleaseNotesTmpl)
		require.Equal(t, header, ctx.ReleaseHeaderTmpl)
		require.Equal(t, footer, ctx.ReleaseFooterTmpl)
	})

	t.Run("rm dist", func(t *testing.T) {
		require.True(t, setup(releaseOpts{
			rmDist: true,
		}).RmDist)
	})
}
