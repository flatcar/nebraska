package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
)

const sampleApplicationID = "b6458005-8f40-4627-b33b-be70a718c48e"

// A renamed application still owns the same instances, so the next tick
// should report the same counts under the new name and stop reporting them
// under the old one. Nothing about the label combinations themselves is
// invalid, they just no longer match what the query returns.
func TestCalculateMetricsDropsStaleLabelSeries(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	defer a.Close()

	require.NoError(t, calculateMetrics(a))

	before := `
# HELP nebraska_application_instances_per_channel Number of applications from specific channel running on instances
# TYPE nebraska_application_instances_per_channel gauge
nebraska_application_instances_per_channel{application="Sample application",channel="Failing",version="1.0.1"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Master",version="1.0.1"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Master",version="1.0.2"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Master",version="1.0.3"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Master",version="1.0.4"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Stable",version="1.0.1"} 1
nebraska_application_instances_per_channel{application="Sample application",channel="Stable",version="1.0.2"} 2
nebraska_application_instances_per_channel{application="Sample application",channel="Stable",version="1.0.3"} 4
`
	require.NoError(t, testutil.CollectAndCompare(appInstancePerChannelGaugeMetric, strings.NewReader(before)))

	adminSvc := admin.NewService(a.Reads())
	app, err := a.Reads().GetApp(sampleApplicationID)
	require.NoError(t, err)
	app.Name = "Sample application renamed"
	require.NoError(t, adminSvc.UpdateApp(app))

	require.NoError(t, calculateMetrics(a))

	// Same 8 lines, only the application name changed. If the old name is
	// still present too, CollectAndCompare fails because the exposed set no
	// longer matches exactly this one.
	after := strings.ReplaceAll(before, "Sample application", "Sample application renamed")
	require.NoError(t, testutil.CollectAndCompare(appInstancePerChannelGaugeMetric, strings.NewReader(after)),
		"the old application name must not still be exposed after a rename")
}
