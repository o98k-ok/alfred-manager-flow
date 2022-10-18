package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/o98k-ok/lazy/v2/alfred"
	"github.com/o98k-ok/lazy/v2/mac"
)

type Flow struct {
	Name    string
	Desc    string
	WebSite string
	User    string

	Path    string
	Keyword []string
	Envs    map[string]string
}

func listAllFlows(p string) ([]string, error) {
	var res []string
	dirs, err := ioutil.ReadDir(p)
	if err != nil {
		return res, err
	}

	for _, d := range dirs {
		res = append(res, path.Join(p, d.Name()))
	}
	return res, nil
}

func keywordFormat(v interface{}) string {
	defer func() {
		recover()
	}()

	return v.(map[string]interface{})["config"].(map[string]interface{})["keyword"].(string)
}

func flowDetail(reader io.ReadSeeker) (*Flow, error) {
	var flow Flow

	flow.Envs, _ = alfred.FlowVariablesWithReader(reader)
	_, _ = reader.Seek(0, io.SeekStart)

	fn := func(keys []string, handler func(interface{})) error {
		res, e := mac.DefaultsRead(reader, keys)
		if e != nil {
			return e
		}
		_, _ = reader.Seek(0, io.SeekStart)
		handler(res)
		return nil
	}

	binds := map[string]func(interface{}){
		"name":        func(i interface{}) { flow.Name = i.(string) },
		"description": func(i interface{}) { flow.Desc = i.(string) },
		"createdby":   func(i interface{}) { flow.User = i.(string) },
		"webaddress":  func(i interface{}) { flow.WebSite = i.(string) },
		"objects": func(i interface{}) {
			val, ok := i.([]interface{})
			if !ok {
				return
			}

			for _, v := range val {
				k := keywordFormat(v)
				if len(k) != 0 {
					flow.Keyword = append(flow.Keyword, k)
				}
			}
		},
	}

	for k, v := range binds {
		e := fn([]string{k}, v)
		if e != nil {
			return &flow, e
		}
	}
	return &flow, nil
}

func searchFlows(flows []*Flow, key string) []*Flow {
	var result []*Flow
	for _, f := range flows {
		if strings.Contains(strings.ToLower(fmt.Sprint(f)), strings.ToLower(key)) {
			result = append(result, f)
		}
	}
	return result
}

func GetWorkflows(items []string) {
	var flows []*Flow

	// root := "/Users/shadow/Library/Mobile Documents/com~apple~CloudDocs/Downloads/alfred/Alfred.alfredpreferences/workflows/"
	root := os.Getenv("alfred_preferences") + "/workflows"
	paths, _ := listAllFlows(root)

	if len(paths) == 0 {
		alfred.Log("not exsit workflows in path %v", root)
	}

	for _, flow := range paths {
		reader, err := os.Open(path.Join(flow, "info.plist"))
		if err != nil {
			alfred.Deliver(alfred.ErrItems("open info.plist error", err).Encode())
			return
		}
		defer reader.Close()

		f, err := flowDetail(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get flow %s detail failed %v", flow, err)
			continue
		}

		f.Path = flow
		flows = append(flows, f)
	}

	// according to keyword for searching
	searched := flows
	for _, item := range items {
		searched = searchFlows(searched, item)
	}

	// to alfred struct
	var alfreds alfred.Items
	for _, f := range searched {
		passing, err := json.Marshal(f)
		if err != nil {
			alfred.Log("encode args error %v", err)
			continue
		}
		alfreds.Append(alfred.NewItem(f.Name, f.Desc, string(passing)))
	}
	alfred.Deliver(alfreds.Encode())
}
