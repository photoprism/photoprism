package clusters

import "errors"

var (
	errEmptySet       = errors.New("Empty training set")
	errNotTrained     = errors.New("You need to train the algorithm first")
	errZeroIterations = errors.New("Number of iterations cannot be less than 1")
	errOneCluster     = errors.New("Number of clusters cannot be less than 2")
	errZeroEpsilon    = errors.New("Epsilon cannot be 0")
	errZeroMinpts     = errors.New("MinPts cannot be 0")
	errZeroWorkers    = errors.New("Number of workers cannot be less than 0")
	errZeroXi         = errors.New("Xi cannot be 0")
	errInvalidRange   = errors.New("Range is invalid")
)
