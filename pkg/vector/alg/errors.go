package alg

import "errors"

var (
	errEmptySet       = errors.New("empty training set")
	errZeroIterations = errors.New("number of iterations cannot be less than 1")
	errOneCluster     = errors.New("number of clusters cannot be less than 2")
	errZeroEpsilon    = errors.New("epsilon cannot be 0")
	errZeroMinpts     = errors.New("minpts cannot be 0")
	errZeroWorkers    = errors.New("number of workers cannot be less than 0")
	errZeroXi         = errors.New("xi cannot be 0")
	errInvalidRange   = errors.New("range is invalid")
)
