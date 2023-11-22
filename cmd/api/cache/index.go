package cache

import (
	"sort"
	"strconv"
	"strings"

	"github.com/crypto_pickle/internal/s3_client"
)

type IndexElement struct {
	key        string
	format     string
	start      int
	end        int
	downloaded bool
}

type Index []IndexElement

// Sorting Interface

func (index Index) Len() int {
	return len(index)
}

func (index Index) Swap(i, j int) {
	index[i], index[j] = index[j], index[i]
}

func (index Index) Less(i, j int) bool {
	return index[i].end < index[i].start
}

// General Functions

func NewIndex(client *s3_client.S3Client, symbol string) Index {
	fileList := client.ListObjects("datapickles", symbol)

	newIndex := make(Index, len(fileList))
	for i, s := range fileList {
		withoutSym := strings.Split(s, "/")[1]
		withoutFormat := strings.Split(withoutSym, ".")

		format := withoutFormat[1]
		times := strings.Split(withoutFormat[0], "-")

		newIndex[i].key = withoutSym
		newIndex[i].format = format
		newIndex[i].start, _ = strconv.Atoi(times[0])
		newIndex[i].end, _ = strconv.Atoi(times[1])
		newIndex[i].downloaded = false
	}

	sort.Sort(newIndex)

	return newIndex
}

func TransferElements(old_index *Index, new_index *Index) {
	i, j := 0, 0
	for i < len(*new_index) && j < len(*old_index) {
		if (*new_index)[i].start < (*old_index)[j].start {
			i++
		} else if (*new_index)[i].start > (*old_index)[j].start {
			j++
		} else {
			(*new_index)[i].downloaded = (*old_index)[i].downloaded

			i++
			j++
		}
	}
}

func (index Index) FindKey(t int) (*IndexElement, int) {
	for i := 0; i < len(index); i++ {
		if index[i].start <= t && index[i].end >= t {
			return &index[i], i
		}
	}

	return nil, -1
}

func (index Index) GetNext(i int) *IndexElement {
	if i+1 < len(index) {
		return &index[i+1]
	}
	return nil
}

func (index Index) GetEarliestTime() int {
	return index[0].start + 100
}

func (index Index) GetLatestTime() int {
	return index[len(index)-1].end - 100
}
