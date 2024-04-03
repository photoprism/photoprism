package config

import (
	"errors"
	"math"

	"github.com/DataIntelligenceCrew/go-faiss"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const metricInnerProduct = 0
const photoEmbeddingsDimension = 512

func normL2sqr(arr []float32) float32 {
	sum := float32(0)
	for _, val := range arr {
		sum += val * val
	}
	return sum
}

func renormL2(arr []float32) []float32 {
	if len(arr) == 0 {
		return []float32{}
	}

	sum := normL2sqr(arr)
	if sum == 0 {
		return make([]float32, len(arr))
	}

	result := make([]float32, len(arr))
	sqrtSum := float32(math.Sqrt(float64(sum)))
	for i, val := range arr {
		result[i] = val / sqrtSum
	}
	return result
}

func (c *Config) initEmbeddingsIndex() error {
	index, err := faiss.IndexFactory(photoEmbeddingsDimension, "IDMap,Flat", metricInnerProduct)
	if err != nil {
		return err
	}
	c.embeddingsIndex = index

	return nil
}

func (c *Config) AddToEmbeddingsIndex(id int64, x []float32) error {
	if c.embeddingsIndex == nil {
		return errors.New("Photos embeddings index is not initialized")
	}
	if len(x) != photoEmbeddingsDimension {
		return errors.New("Wrong vector dimensions")
	}
	selector, err := faiss.NewIDSelectorBatch([]int64{id})
	_, err = c.embeddingsIndex.RemoveIDs(selector)
	if err != nil {
		return err
	}
	err = c.embeddingsIndex.AddWithIDs(renormL2(x), []int64{id})
	if err != nil {
		return err
	}
	log.Debugf("embeddings: index size %d", c.embeddingsIndex.Ntotal())
	return err
}

func (c *Config) SearchEmbeddingsIndex(x []float32, k int64) ([]float32, []int64, error) {
	if c.embeddingsIndex == nil {
		return nil, nil, errors.New("Photos embeddings index is not initialized")
	}
	if !c.embeddingsIndex.IsTrained() {
		return nil, nil, errors.New("Photos embeddings index is not trained")
	}
	if len(x) != photoEmbeddingsDimension {
		return nil, nil, errors.New("Wrong query dimensions")
	}
	dist, labels, err := c.embeddingsIndex.Search(renormL2(x), k)
	if err != nil {
		return dist, labels, err
	}
	return dist, labels, nil
}
