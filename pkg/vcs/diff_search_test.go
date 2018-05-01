	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
		"git tag mytag HEAD",
						Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
						Refs:       []string{"refs/heads/master", "refs/tags/mytag"},

				// With path exclude/include filters
				{
					Paths: vcs.PathOptions{
						IncludePatterns: []string{"g"},
						ExcludePattern:  "f",
						IsRegExp:        true,
					},
				}: nil, // empty
			},
		},
	}

	for label, test := range tests {
		for opt, want := range test.want {
			results, complete, err := test.repo.RawLogDiffSearch(ctx, *opt)
			if err != nil {
				t.Errorf("%s: %+v: %s", label, *opt, err)
				continue
			}
			if !complete {
				t.Errorf("%s: !complete", label)
			}
			for _, r := range results {
				r.DiffHighlights = nil // Highlights is tested separately
			}
			if !reflect.DeepEqual(results, want) {
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, asJSON(results), asJSON(want))
			}
		}
	}
}

func TestRepository_RawLogDiffSearch_emptyCommit(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m empty --allow-empty --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo interface {
			RawLogDiffSearch(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, bool, error)
		}
		want map[*vcs.RawLogDiffSearchOptions][]*vcs.LogCommitSearchResult
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			want: map[*vcs.RawLogDiffSearchOptions][]*vcs.LogCommitSearchResult{
				{
					Paths: vcs.PathOptions{IncludePatterns: []string{"/xyz.txt"}, IsRegExp: true},
				}: nil, // want no matches