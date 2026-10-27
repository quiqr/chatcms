package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/publish"
)

// waitingInSites answers how much of a project is waiting to be published,
// counting the same things the publish screen lists. It lives here because
// tenancy knows who may reach what and publish knows what has changed;
// neither should have to know the other.
//
// It counts changes rather than versions on purpose: a person reading
// "Publish 2" beside "Ready to publish 2 changes" expects one number, and
// counting versions gave them two different ones.
type waitingInSites struct {
	sources publishSources
}

func (w waitingInSites) UnpublishedIn(ctx context.Context, projectID string) (int, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return 0, nil
	}

	site, done, err := w.sources.Site(ctx, id)
	if err != nil {
		// A project with no site yet has nothing waiting, which is not a
		// failure worth refusing the screen that lists it.
		return 0, nil
	}
	defer done()

	m, err := site.Model()
	if err != nil {
		m = nil
	}

	summary, err := publish.Describe(ctx, site, m)
	if err != nil {
		return 0, err
	}
	return summary.Count(), nil
}
