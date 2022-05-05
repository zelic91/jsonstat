package main

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

type statKey struct {
	Key  string
	Hash string
}

type statValue struct {
	Count int
	Size  int
}

type result struct {
	Key   string
	Hash  string
	Size  int
	Count int
	Total int
}

var stats map[statKey]statValue

func init() {
	stats = map[statKey]statValue{}
}

func main() {
	if len(os.Args) == 1 {
		log.Panic("You must provide file name.")
	}

	fileName := os.Args[1]

	if len(fileName) == 0 {
		log.Panic("Invalid file name.")
	}

	file, err := ioutil.ReadFile(fileName)

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
		checksum := fmt.Sprintf("%x", sha512.Sum512(bytes))

		sk := statKey{
			Key:  lastKey,
			Hash: checksum,
		}

		s, ok := stats[sk]
		if !ok {
			s = statValue{
				Count: 1,
				Size:  len(bytes),
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
	data := []result{}
	for key, value := range stats {
		data = append(data, result{
			Key:   key.Key,
			Hash:  key.Hash,
			Size:  value.Size,
			Count: value.Count,
			Total: value.Size * value.Count,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Count > data[j].Count
	})

	data = data[0:20]

	tableData := [][]string{}

	for _, value := range data {
		tableData = append(tableData, []string{
			value.Key,
			value.Hash[len(value.Hash)-10:],
			fmt.Sprintf("%d", value.Count),
			humanize.Bytes(uint64(value.Size)),
			humanize.Bytes(uint64(value.Total)),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Checksum", "Count", "Unit Size", "Total"})

	for _, v := range tableData {
		table.Append(v)
	}

	table.Render()
}
