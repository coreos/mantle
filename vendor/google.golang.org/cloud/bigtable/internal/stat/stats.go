// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stat

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"text/tabwriter"
	"time"
)

type byDuration []time.Duration

func (data byDuration) Len() int           { return len(data) }
func (data byDuration) Swap(i, j int)      { data[i], data[j] = data[j], data[i] }
func (data byDuration) Less(i, j int) bool { return data[i] < data[j] }

// quantile returns a value representing the kth of q quantiles.
// May alter the order of data.
func quantile(data []time.Duration, k, q int) (quantile time.Duration, ok bool) {
	if len(data) < 1 {
		return 0, false
	}
	if k > q {
		return 0, false
	}
	if k < 0 || q < 1 {
		return 0, false
	}

	sort.Sort(byDuration(data))

	if k == 0 {
		return data[0], true
	}
	if k == q {
		return data[len(data)-1], true
	}

	bucketSize := float64(len(data)-1) / float64(q)
	i := float64(k) * bucketSize

	lower := int(math.Trunc(i))
	var upper int
	if i > float64(lower) && lower+1 < len(data) {
		// If the quantile lies between two elements
		upper = lower + 1
	} else {
		upper = lower
	}
	weightUpper := i - float64(lower)
	weightLower := 1 - weightUpper
	return time.Duration(weightLower*float64(data[lower]) + weightUpper*float64(data[upper])), true
}

type Aggregate struct {
	Min, Median, Max time.Duration
	P95, P99         time.Duration // percentiles
}

// NewAggregate constructs an aggregate from latencies. Returns nil if latencies does not contain aggregateable data.
func NewAggregate(latencies []time.Duration) *Aggregate {
	var agg Aggregate

	if len(latencies) == 0 {
		return nil
	}
	var ok bool
	if agg.Min, ok = quantile(latencies, 0, 2); !ok {
		return nil
	}
	if agg.Median, ok = quantile(latencies, 1, 2); !ok {
		return nil
	}
	if agg.Max, ok = quantile(latencies, 2, 2); !ok {
		return nil
	}
	if agg.P95, ok = quantile(latencies, 95, 100); !ok {
		return nil
	}
	if agg.P99, ok = quantile(latencies, 99, 100); !ok {
		return nil
	}
	return &agg
}

func (agg *Aggregate) String() string {
	if agg == nil {
		return "no data"
	}
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0) // one-space padding
	fmt.Fprintf(tw, "min:\t%v\nmedian:\t%v\nmax:\t%v\n95th percentile:\t%v\n99th percentile:\t%v\n",
		agg.Min, agg.Median, agg.Max, agg.P95, agg.P99)
	tw.Flush()
	return buf.String()
}
