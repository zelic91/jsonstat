package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"github.com/olekukonko/tablewriter"
)

type statKey struct {
	Key  string
	Hash string
}

type statValue struct {
	Count int
}

var stats map[statKey]statValue

func init() {
	stats = map[statKey]statValue{}
}

func main() {
	file, err := ioutil.ReadFile("./getListViewData.json")

	if err != nil {
		log.Panicf("Cannot open file: %v", err)
	}

	var rawData map[string]interface{}

	err = json.Unmarshal(file, &rawData)
	if err != nil {
		log.Panicf("Cannot parse JSON content: %v", err)
	}

	path := []string{}

	// Main processing stuff
	doNode(path, rawData)

	// Render result as a table
	printResult()
}

func doNode(path []string, root interface{}) {
	kind := reflect.ValueOf(root).Kind()

	if (kind != reflect.Map) && (kind != reflect.Slice) {
		// fmt.Printf("Terminate with kind %v\n", kind)
		return
	}

	if len(path) > 0 {
		lastKey := path[len(path)-1]

		bytes, err := json.Marshal(root)
		if err != nil {
			log.Panicf("Cannot build JSON byte: %v", err)
		}
		checksum := fmt.Sprintf("%x", md5.Sum(bytes))

		sk := statKey{
			Key:  lastKey,
			Hash: checksum,
		}

		s, ok := stats[sk]
		if !ok {
			s = statValue{
				Count: 1,
			}
		} else {
			s.Count += 1
		}

		stats[sk] = s
	}

	if kind == reflect.Map {
		for key, value := range root.(map[string]interface{}) {
			temp := append(path, key)
			doNode(temp, value)
		}
	} else if kind == reflect.Slice {
		for index, value := range root.([]interface{}) {
			temp := append(path, fmt.Sprintf("%d", index))
			doNode(temp, value)
		}
	}

}

func printResult() {
	data := [][]string{}

	for key, value := range stats {
		data = append(data, []string{
			key.Key,
			key.Hash,
			fmt.Sprintf("%d", value.Count),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Checksum", "Count"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}
