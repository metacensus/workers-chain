// Package charts ships the Helm charts with the worker that renders them.
package charts

import "embed"

// FS holds the charts at fabric/<chart>.
//
//go:embed all:fabric
var FS embed.FS
