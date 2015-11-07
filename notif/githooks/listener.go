package githooks

import (
	"github.com/AaronO/go-git-http"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
)

func init() {
	events.RegisterListener(&gitHookListener{})
}

type gitHookListener struct{}

func (g *gitHookListener) Scopes() []string {
	return []string{"app:githooks"}
}

func (g *gitHookListener) Start(ctx context.Context) {
	notifyCallback := func(id events.EventID, p notif.GitPayload) {
		notifyGitEvent(ctx, id, p)
	}
	buildCallback := func(id events.EventID, p notif.GitPayload) {
		buildHook(ctx, id, p)
	}

	events.Subscribe(GitPushEvent, notifyCallback)
	events.Subscribe(GitCreateEvent, notifyCallback)
	events.Subscribe(GitDeleteEvent, notifyCallback)

	events.Subscribe(GitPushEvent, buildCallback)
}

func notifyGitEvent(ctx context.Context, id events.EventID, payload notif.GitPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Warn("postPushHook: error getting user", "error", err)
		return
	}

	repo := payload.Repo
	event := payload.Event
	branchURL, err := router.Rel.URLToRepoRev(repo.URI, event.Branch)
	if err != nil {
		log15.Warn("postPushHook: error resolving branch URL", "repo", repo.URI, "branch", event.Branch, "error", err)
		return
	}

	absBranchURL := conf.AppURL(ctx).ResolveReference(branchURL).String()

	if id == GitCreateEvent {
		cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
			Actor:       payload.Actor,
			ActionType:  "created the branch",
			ObjectURL:   absBranchURL,
			ObjectRepo:  repo.URI+"@"+event.Branch,
		})
		return
	}

	if id == GitDeleteEvent {
		cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
			Actor:       payload.Actor,
			ActionType:  "deleted the branch",
			ObjectURL:   absBranchURL,
			ObjectRepo:  repo.URI+"@"+event.Branch,
		})
		return
	}

	// See how many commits were pushed.
	commits, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: repo,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Head:         event.Commit,
			Base:         event.Last,
			RefreshCache: true,
			ListOptions:  sourcegraph.ListOptions{PerPage: 1000},
		},
	})
	if err != nil {
		log15.Warn("postPushHook: error fetching push commits", "error", err)
		commits = &sourcegraph.CommitList{}
	}

	var commitsNoun string
	if len(commits.Commits) == 1 {
		commitsNoun = "commit"
	} else {
		commitsNoun = "commits"
	}
	var commitMessages []string
	for i, c := range commits.Commits {
		if i > 10 {
			break
		}
		commitURL := router.Rel.URLToRepoCommit(repo.URI, string(c.ID))
		commitMessages = append(commitMessages, fmt.Sprintf("<%s|%s>: %s",
			conf.AppURL(ctx).ResolveReference(commitURL).String(),
			c.ID[:6],
			textutil.ShortCommitMessage(80, c.Message),
		))
	}

	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:       payload.Actor,
		ActionType:  fmt.Sprintf("pushed *%d %s* to", len(commits.Commits), commitsNoun),
		ObjectURL:   absBranchURL,
		ObjectRepo:  repo.URI+"@"+event.Branch,
		ActionContent: strings.Join(commitMessages, "\n"),
	})
}

func buildHook(ctx context.Context, id events.EventID, payload notif.GitPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)
	repo := payload.Repo
	event := payload.Event
	if event.Type == githttp.PUSH || event.Type == githttp.PUSH_FORCE {
		_, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{
			RepoRev: sourcegraph.RepoRevSpec{RepoSpec: repo, Rev: event.Branch, CommitID: event.Commit},
			Opt:     &sourcegraph.BuildCreateOptions{BuildConfig: sourcegraph.BuildConfig{Import: true, Queue: true}},
		})
		if err != nil {
			log15.Warn("postPushHook: failed to create build", "err", err, "repo", repo.URI, "commit", event.Commit, "branch", event.Branch)
			return
		}
		log15.Debug("postPushHook: build created", "repo", repo.URI, "branch", event.Branch, "commit", event.Commit)
	}
}
