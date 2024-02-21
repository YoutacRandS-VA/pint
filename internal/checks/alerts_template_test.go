package checks_test

import (
	"fmt"
	"testing"

	"github.com/cloudflare/pint/internal/checks"
	"github.com/cloudflare/pint/internal/parser"
	"github.com/cloudflare/pint/internal/promapi"
)

func newTemplateCheck(_ *promapi.FailoverGroup) checks.RuleChecker {
	return checks.NewTemplateCheck()
}

func humanizeText(call string) string {
	return fmt.Sprintf("Using the value of `%s` inside this annotation might be hard to read, consider using one of humanize template functions to make it more human friendly.", call)
}

func TestTemplateCheck(t *testing.T) {
	testCases := []checkTest{
		{
			description: "skips recording rule",
			content:     "- record: foo\n  expr: sum(foo)\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "invalid syntax in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: 'Instance {{ $label.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `undefined variable \"$label\"`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "invalid function in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: '{{ $value | xxx }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `function \"xxx\" not defined`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "valid syntax in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: 'Instance {{ $labels.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "invalid syntax in labels",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: 'Instance {{ $label.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `undefined variable \"$label\"`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "invalid function in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: '{{ $value | xxx }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `function \"xxx\" not defined`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "valid syntax in labels",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: 'Instance {{ $labels.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "{{ $value}} in label key",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    '{{ $value}}': bar\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "{{$value}} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{$value}}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{$value}} in multiple labels",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: '{{ .Value }}'\n    baz: '{{$value}}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `.Value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{  $value  }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: |\n      foo is {{  $value | humanizePercentage }}%\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  6,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{  $value  }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: |\n      foo is {{$value|humanizePercentage}}%\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  6,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ .Value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: 'value {{ .Value }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `.Value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ .Value|humanize }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ .Value|humanize }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `.Value` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ $foo := $value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ $foo := $value }}{{ $foo }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$foo` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ $foo := .Value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ $foo := .Value }}{{ $foo }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Using `$foo` in labels will generate a new alert on every value change, move it to annotations.",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (by)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) > 0\n  annotations:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (by)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) > 0\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  annotations:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  labels:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (or)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) by(job) or sum(bar)\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (1+)",
			content:     "- alert: Foo Is Down\n  expr: 1 + sum(foo) by(notjob)\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (group_left)",
			content: `
- alert: Foo Is Down
  expr: count(build_info) by (instance, version) != ignoring(package) group_left(foo) count(package_installed) by (instance, version, package)
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.foo }} is down'
    help: '{{ $labels.ixtance }}'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 6,
							Last:  6,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `ixtance` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "don't trigger for label_replace() provided labels",
			content: `
- alert: label_replace_not_checked_correctly
  expr: |
    label_replace(
      sum by (pod) (pod_status) > 0
      ,"cluster", "$1", "pod", "(.*)"
    )
  annotations:
    summary: "Some error found in {{ $labels.cluster }}"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label present on metrics (absent)",
			content: `
- alert: Foo Is Missing
  expr: absent(foo{job="bar", instance="server1"})
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label missing from metrics (absent)",
			content: `
- alert: Foo Is Missing
  expr: absent(foo{job="bar"}) AND on(job) foo
  labels:
    instance: '{{ $labels.instance }}'
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.foo }} is missing'
    help: '{{ $labels.xxx }}'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 7,
							Last:  7,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 7,
							Last:  7,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `foo` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 8,
							Last:  8,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `xxx` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label present on metrics (absent(sum))",
			content: `
- alert: Foo Is Missing
  expr: absent(sum(foo) by(job, instance))
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label missing from metrics (absent(sum))",
			content: `
- alert: Foo Is Missing
  expr: absent(sum(foo) by(job))
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but the query removes it.",
						Details:  checks.TemplateCheckAggregationDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (absent({job=~}))",
			content: `
- alert: Foo Is Missing
  expr: absent({job=~".+"})
  annotations:
    summary: '{{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (absent()) / multiple",
			content: `
- alert: Foo Is Missing
  expr: absent(foo) or absent(bar)
  annotations:
    summary: '{{ .Labels.job }} / {{$labels.job}} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but `absent()` is not passing it.",
						Details:  checks.TemplateCheckAbsentDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "absent() * on() group_left(...) foo",
			content: `
- alert: Foo
  expr: absent(foo{job="xxx"}) * on() group_left(cluster, env) bar
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "absent() * on() group_left() bar",
			content: `
- alert: Foo
  expr: absent(foo{job="xxx"}) * on() group_left() bar
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "bar * on() group_right(...) absent()",
			content: `
- alert: Foo
  expr: bar * on() group_right(cluster, env) absent(foo{job="xxx"})
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "bar * on() group_right() absent()",
			content: `
- alert: Foo
  expr: bar * on() group_right() absent(foo{job="xxx"})
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "foo and on() absent(bar)",
			content: `
- alert: Foo
  expr: foo and on() absent(bar)
  annotations:
    summary: '{{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "no humanize on rate()",
			content: `
- alert: Foo
  expr: rate(errors[2m]) > 0
  annotations:
    summary: "Seeing {{ $value }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     humanizeText("rate(errors[2m])"),
						Severity: checks.Information,
					},
				}
			},
		},
		{
			description: "no humanize on rate() / alias",
			content: `
- alert: Foo
  expr: rate(errors[2m]) > 0
  annotations:
    summary: "{{ $foo := $value }}{{ $bar := $foo }} Seeing {{ $bar }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     humanizeText("rate(errors[2m])"),
						Severity: checks.Information,
					},
				}
			},
		},
		{
			description: "no humanize on irate()",
			content: `
- alert: Foo
  expr: irate(errors[2m]) > 0
  annotations:
    summary: "Seeing {{ .Value }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     humanizeText("irate(errors[2m])"),
						Severity: checks.Information,
					},
				}
			},
		},
		{
			description: "no humanize on irate()",
			content: `
- alert: Foo
  expr: deriv(errors[2m]) > 0
  annotations:
    summary: "Seeing {{ .Value }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     humanizeText("deriv(errors[2m])"),
						Severity: checks.Information,
					},
				}
			},
		},
		{
			description: "rate() but no $value",
			content: `
- alert: Foo
  expr: rate(errors[2m]) > 0
  annotations:
    summary: "Seeing errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanize passed to value",
			content: `
- alert: Foo
  expr: rate(errors[2m]) > 0
  annotations:
    summary: "Seeing {{ $value | humanize }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanizePercentage passed to value",
			content: `
- alert: Foo
  expr: (sum(rate(errors[2m])) / sum(rate(requests[2m]))) > 0.1
  annotations:
    summary: "Seeing {{ $value | humanizePercentage }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanizeDuration passed to value",
			content: `
- alert: Foo
  expr: (sum(rate(errors[2m])) / sum(rate(requests[2m]))) > 0.1
  annotations:
    summary: "Seeing {{ $value | humanizeDuration }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanize not needed on count()",
			content: `
- alert: Foo
  expr: count(rate(errors[2m]) > 0) > 0
  annotations:
    summary: "Seeing {{ $value }} instances with errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanize not needed on rate() used in RHS",
			content: `
- alert: Foo
  expr: foo > on() sum(rate(errors[2m])
  annotations:
    summary: "Seeing {{ $value }} instances with errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "humanize not needed on round(rate())",
			content: `
- alert: Foo
  expr: round(rate(errors_total[5m]), 1) > 0
  annotations:
    summary: "Seeing {{ $value }} instances with errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "toTime",
			content: `
- alert: Foo
  expr: up == 0
  annotations:
    summary: "{{ $value | toTime }}"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "template query with syntax error",
			content: `
- alert: Foo
  expr: up == 0
  annotations:
    summary: |
      {{ with printf "sum({job='%s'}) by(" .Labels.job | query }}
      {{ . | first | label "instance" }}
      {{ end }}
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  8,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `163: executing \"summary\" at <query>: error calling query: 1:18: parse error: unclosed left parenthesis`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "template query with bogus function",
			content: `
- alert: Foo
  expr: up == 0
  annotations:
    summary: |
      {{ with printf "suz({job='%s'})" .Labels.job | query }}
      {{ . | first | label "instance" }}
      {{ end }}
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  8,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `159: executing \"summary\" at <query>: error calling query: 1:1: parse error: unknown function with name \"suz\"`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "$value | first",
			content: `
- alert: Foo
  expr: rate(errors[2m])
  annotations:
    summary: "{{ $value | first }} errors"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `124: executing \"summary\" at <first>: wrong type for value; expected template.queryResult; got float64`.",
						Severity: checks.Fatal,
						Details:  checks.TemplateCheckSyntaxDetails,
					},
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  5,
						},
						Reporter: checks.TemplateCheckName,
						Text:     humanizeText("rate(errors[2m])"),
						Severity: checks.Information,
					},
				}
			},
		},
		{
			description: "template query with with bogus range",
			content: `
- alert: Foo
  expr: up == 0
  annotations:
    summary: |
      {{ range query "up xxx" }}
      {{ .Labels.instance }} {{ .Value }}
      {{ end }}
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 5,
							Last:  8,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template failed to parse with this error: `121: executing \"summary\" at <query \"up xxx\">: error calling query: 1:4: parse error: unexpected identifier \"xxx\"`.",
						Details:  checks.TemplateCheckSyntaxDetails,
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "template query with valid expr",
			content: `
- alert: Foo
  expr: up{job="bar"} == 0
  annotations:
    summary: Instance {{ printf "up{job='bar', instance='%s'}" $labels.instance | query | first | value }} is down'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		/*
					TODO
					{
						description: "template query removes instance",
						content: `
			- alert: Foo
			  expr: up == 0
			  annotations:
			    summary: |
			      {{ with printf "sum({job='%s'})" .Labels.job | query }}
			      {{ . | first | label "instance" }}
			      {{ end }}
			`,
						checker:    newTemplateCheck,
						prometheus: noProm,
						problems: func(_ string) []checks.Problem {
							return []checks.Problem{
								{
												    {{ with printf "sum({job='%s'})" .Labels.job | query }}
			    {{ . | first | label "instance" }}`,
									Lines: parser.LineRange{
						First: 5,
						Last:  8,
					},
									Reporter: checks.TemplateCheckName,
									Text:     `"summary" annotation template sends a query that is using "instance" label but that query removes it`,
									Severity: checks.Bug,
								},
							}
						},
					},
		*/
		{
			description: "sub aggregation",
			content: `
- alert: Foo
  expr: |
    (
      sum(foo:sum > 0) without(notify)
      * on(job) group_left(notify)
      job:notify
    )
    and on(job)
    sum(foo:count) by(job) > 20
  labels:
    notify: "{{ $labels.notify }}"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "abs / scalar",
			content: `
- alert: ScyllaNonBalancedcqlTraffic
  expr: >
    abs(rate(scylla_cql_updates{conditional="no"}[1m]) - scalar(avg(rate(scylla_cql_updates{conditional="no"}[1m]))))
    /
    scalar(stddev(rate(scylla_cql_updates{conditional="no"}[1m])) + 100) > 2
  for: 10s
  labels:
    advisor: balanced
    dashboard: cql
    severity: moderate
    status: "1"
    team: team_devops
  annotations:
    description: CQL queries are not balanced among shards {{ $labels.instance }} shard {{ $labels.shard }}
    summary: CQL queries are not balanced
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label from vector(0)",
			content:     "- alert: DeadMansSwitch\n  expr: vector(1)\n  annotations:\n    summary: 'Deadmans switch on {{ $labels.instance }} is firing'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but the query doesn't produce any labels.",
						Details:  checks.TemplateCheckLabelsDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "labels label from vector(0)",
			content:     "- alert: DeadMansSwitch\n  expr: vector(1)\n  labels:\n    summary: 'Deadmans switch on {{ $labels.instance }} is firing'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but the query doesn't produce any labels.",
						Details:  checks.TemplateCheckLabelsDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label from number",
			content:     "- alert: DeadMansSwitch\n  expr: 1 > bool 0\n  annotations:\n    summary: 'Deadmans switch on {{ $labels.instance }} / {{ $labels.job }} is firing'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `instance` label but the query doesn't produce any labels.",
						Details:  checks.TemplateCheckLabelsDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job` label but the query doesn't produce any labels.",
						Details:  checks.TemplateCheckLabelsDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "foo / on(...) bar",
			content: `- alert: Foo
  expr: container_file_descriptors / on (instance, app_name) container_ulimits_soft{ulimit="max_open_files"}
  annotations:
    summary: "{{ $labels.app_type }} is using {{ $value }} fds."
  labels:
    job: "{{ $labels.job_name }}"
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(_ string) []checks.Problem {
				return []checks.Problem{
					{
						Lines: parser.LineRange{
							First: 6,
							Last:  6,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `job_name` label but the query uses `on(...)` without it being set there, this label will be missing from the query result.",
						Details:  checks.TemplateCheckOnDetails,
						Severity: checks.Bug,
					},
					{
						Lines: parser.LineRange{
							First: 4,
							Last:  4,
						},
						Reporter: checks.TemplateCheckName,
						Text:     "Template is using `app_type` label but the query uses `on(...)` without it being set there, this label will be missing from the query result.",
						Details:  checks.TemplateCheckOnDetails,
						Severity: checks.Bug,
					},
				}
			},
		},
	}
	runTests(t, testCases)
}
