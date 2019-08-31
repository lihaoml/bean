// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package distuv

import (
	"sort"
	"testing"
)

func TestLognormal(t *testing.T) {
	for i, dist := range []LogNormal{
		{
			Mu:    0.1,
			Sigma: 0.3,
		},
		{
			Mu:    0.01,
			Sigma: 0.01,
		},
		{
			Mu:    2,
			Sigma: 0.01,
		},
	} {
		f := dist
		tol := 1e-2
		const n = 1e5
		x := make([]float64, n)
		generateSamples(x, f)
		sort.Float64s(x)

		checkMean(t, i, x, f, tol)
		checkVarAndStd(t, i, x, f, tol)
		checkEntropy(t, i, x, f, tol)
		checkExKurtosis(t, i, x, f, 2e-1)
		checkSkewness(t, i, x, f, 5e-2)
		checkMedian(t, i, x, f, tol)
		checkQuantileCDFSurvival(t, i, x, f, tol)
		checkProbContinuous(t, i, x, f, 1e-10)
		checkProbQuantContinuous(t, i, x, f, tol)
	}
}

// See https://github.com/gonum/gonum/issues/577 for details.
func TestLognormalIssue577(t *testing.T) {
	x := 1.0e-16
	max := 1.0e-295
	cdf := LogNormal{Mu: 0, Sigma: 1}.CDF(x)
	if cdf <= 0 {
		t.Errorf("LogNormal{0,1}.CDF(%e) should be positive. got: %e", x, cdf)
	}
	if cdf > max {
		t.Errorf("LogNormal{0,1}.CDF(%e) is greater than %e. got: %e", x, max, cdf)
	}
}
