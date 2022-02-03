package clip

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/tidwall/gjson"
)

// EmbeddingDB talks with a qdrant database using their REST interface
type EmbeddingDB struct {
	Url        string
	Collection string
	VectorSize uint
}

func (db *EmbeddingDB) collectionUrl() string {
	return db.Url + "/collections/" + db.Collection
}
func (db *EmbeddingDB) getStatus() (string, error) {
	resp, err := http.Get(db.collectionUrl())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("could not get collection status! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Printf("status body: %v\n", string(body))
	return gjson.Get(string(body), "status").Str, nil
}

func (db *EmbeddingDB) createCollection() error {
	body := fmt.Sprintf(`{"create_collection": {"distance": "Cosine", "name": "%s", "vector_size": %d}}`, db.Collection, db.VectorSize)
	var jsonData = []byte(body)
	req, err := http.NewRequest("POST", db.Url+"/collections", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("could not create collection! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	return nil
}

// delete collection from embedding db
func (db *EmbeddingDB) DeleteCollection() error {
	body := fmt.Sprintf(`{"delete_collection": "%s"}`, db.Collection)
	fmt.Printf("DeleteCollection body: %s\n", body)
	var jsonData = []byte(body)
	req, err := http.NewRequest("POST", db.Url+"/collections", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("could not delete collection! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	return nil
}

// makes sure that a collection exists
func (db *EmbeddingDB) CreateCollectionIfNotExisting() error {
	status, err := db.getStatus()
	if err != nil || status != "ok" {
		err = db.createCollection()
		if err != nil {
			return err
		}
	}

	status, err = db.getStatus()
	if err != nil {
		return err
	}
	if status != "ok" {
		return fmt.Errorf("collection status not ok after creating! status='%v'", status)
	}
	return nil
}

// save an embedding with an id to embedding database
func (db *EmbeddingDB) SaveEmbedding(embedding face.Embedding, id uint) error {
	body := fmt.Sprintf(`{"upsert_points": {"points": [{"id": %d, "vector": %s}] }}`, id, embedding.MarshalEmbedding())
	fmt.Printf("SaveEmbedding body: %v\n", body)
	var jsonData = []byte(body)
	req, err := http.NewRequest("POST", db.collectionUrl(), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("could not save embedding! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	return nil
}

// Load an embedding by id from embedding database
func (db *EmbeddingDB) LoadEmbedding(id uint) (face.Embedding, error) {
	url := db.Url + fmt.Sprintf("/collections/%s/points/%d", db.Collection, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("could not load embedding! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("LoadEmbedding Response body: %s\n", string(body))
	result := gjson.Get(string(body), "result.vector").Array()
	return face.NewEmbeddingFromJsonArray(result), nil
}

// Delete Embedding by id from embedding database
func (db *EmbeddingDB) DeleteEmbedding(id uint) error {
	body := fmt.Sprintf(`{"delete_points": {"ids": [%d]}}`, id)
	fmt.Printf("DeleteEmbedding body: %v\n", body)
	var jsonData = []byte(body)
	req, err := http.NewRequest("POST", db.collectionUrl(), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("could not delete embedding! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	return nil
}

// get the k neireighst neighbors (maximum n) from an embedding. takes only embeddings with a score > minScore
func (db *EmbeddingDB) KNearestNeighbors(embedding face.Embedding, n uint, minScore float32) ([]uint64, error) {
	body := fmt.Sprintf(`{"top": %d, "vector": %s}`, n, embedding.MarshalEmbedding())
	fmt.Printf("KNearestNeighbors body: %v\n", body)
	var jsonData = []byte(body)
	req, err := http.NewRequest("POST", db.collectionUrl()+"/points/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("could not get KNearestNeighbors! http status='%v' (%v)", resp.StatusCode, resp.Status)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("KNearestNeighbors Response body: %s\n", string(respBody))
	var results = []uint64{}
	jsonResults := gjson.Get(string(respBody), fmt.Sprintf("result.#(score>%f)#.id", minScore)).Array()
	for _, jsonResult := range jsonResults {
		results = append(results, uint64(jsonResult.Num))
	}
	fmt.Printf("json: %v -> results: %v\n", jsonResults, results)
	return results, nil
}
